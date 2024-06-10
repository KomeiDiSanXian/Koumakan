package control

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/bits"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var banCache = make(map[uint64]bool)

// Control controls plugins
type Control[Ctx any] struct {
	Service string
	Cache   map[int64]uint8 // map[gid]status
	Options Option[Ctx]     // options
	Manager *Manager[Ctx]
}

// Enable enables plugin for group
//
// if groupID is 0, it will enable the plugin for all groups
func (c *Control[Ctx]) Enable(groupID int64) {
	var gconf GroupConfig
	c.Manager.RLock()
	err := c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	c.Manager.RUnlock()
	if err != nil {
		gconf.GroupID = groupID
	}
	gconf.Disable = int64(uint64(gconf.Disable) & 0xffffffff_fffffffe) // 变成最近的偶数
	c.Manager.Lock()
	c.Cache[groupID] = 0
	err = c.Manager.DataBase.Insert(c.Service, &gconf)
	c.Manager.Unlock()
	if err != nil {
		log.Errorf("[Control] %s enable failed: %v", c.Service, err)
	}
}

// Disable disables plugin for group
//
// if groupID is 0, it will disable the plugin for all groups
func (c *Control[Ctx]) Disable(groupID int64) {
	var gconf GroupConfig
	c.Manager.RLock()
	err := c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	c.Manager.RUnlock()
	if err != nil {
		gconf.GroupID = groupID
	}
	gconf.Disable = gconf.Disable | 1 // 变成最近的奇数
	c.Manager.Lock()
	c.Cache[groupID] = 1
	err = c.Manager.DataBase.Insert(c.Service, &gconf)
	c.Manager.Unlock()
	if err != nil {
		log.Errorf("[Control] %s disable failed: %v", c.Service, err)
	}
}

// Reset resets the plugin for group
//
// groupID == 0 is not allowed
func (c *Control[Ctx]) Reset(groupID int64) {
	if groupID != 0 {
		c.Manager.Lock()
		if c.Options.DefaultDisable {
			c.Cache[groupID] = 1
		} else {
			c.Cache[groupID] = 0
		}
		err := c.Manager.DataBase.Del(c.Service, "WHERE gid = "+strconv.FormatInt(groupID, 10))
		c.Manager.Unlock()
		if err != nil {
			log.Errorf("[Control] %s reset failed: %v", c.Service, err)
		}
	}
}

// IsEnable checks if the plugin is enabled for group
func (c *Control[Ctx]) IsEnable(groupID int64) bool {
	var gconf GroupConfig
	var err error
	c.Manager.RLock()
	isDisable, ok := c.Cache[0]
	if !ok {
		c.Manager.RLock()
		err = c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = 0")
		c.Manager.RUnlock()
		c.Manager.Lock()
		if err == nil && gconf.GroupID == 0 {
			if gconf.Disable&1 == 0 {
				isDisable = 0
			} else {
				isDisable = 1
			}
		} else {
			isDisable = 2
		}
		c.Cache[0] = isDisable
		ok = true
		c.Manager.Unlock()
		log.Debugf("[Control] cache plugin %s of all: %v", c.Service, isDisable)
	}

	if isDisable != 2 && ok {
		return isDisable == 0
	}

	c.Manager.RLock()
	isDisable, ok = c.Cache[groupID]
	c.Manager.RUnlock()
	if !ok {
		c.Manager.RLock()
		err = c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = "+strconv.FormatInt(groupID, 10))
		c.Manager.RUnlock()
		c.Manager.Lock()
		if err == nil && gconf.GroupID == groupID {
			if gconf.Disable&1 == 0 {
				isDisable = 0
			} else {
				isDisable = 1
			}
		}
		c.Cache[groupID] = isDisable
		ok = true
		c.Manager.Unlock()
		log.Debugf("[Control] cache plugin %s of group %d: %v", c.Service, groupID, isDisable)
	}

	if ok {
		return isDisable == 0
	}

	c.Manager.Lock()
	if c.Options.DefaultDisable {
		isDisable = 1
	} else {
		isDisable = 0
	}
	c.Cache[groupID] = isDisable
	c.Manager.Unlock()
	log.Debugf("[Control] cache plugin %s of group %d(default): %v", c.Service, groupID, isDisable)
	return isDisable == 0
}

// Ban bans plugin permisson for user in group
//
// if groupID is 0, it will ban the plugin for all groups
func (c *Control[Ctx]) Ban(groupID, userID int64) {
	var err error
	var digest [16]byte
	if groupID != 0 {
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", c.Service, userID, groupID)))
		id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
		c.Manager.Lock()
		err = c.Manager.DataBase.Insert(c.Service+"ban", &BanStatus{
			ID:      int64(id),
			UserID:  userID,
			GroupID: groupID,
		})
		banCache[id] = true
		c.Manager.Unlock()

		if err != nil {
			log.Errorf("[Control] %s ban %d in group %d failed: %v", c.Service, userID, groupID, err)
		} else {
			log.Debugf("[Control] %s ban %d in group %d", c.Service, userID, groupID)
		}
		return
	}
	// ban all groups
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", c.Service, userID)))
	id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
	c.Manager.Lock()
	err = c.Manager.DataBase.Insert(c.Service+"ban", &BanStatus{
		ID:      int64(id),
		UserID:  userID,
		GroupID: 0,
	})
	banCache[id] = true
	c.Manager.Unlock()
	if err != nil {
		log.Errorf("[Control] %s ban %d in all groups failed: %v", c.Service, userID, err)
	} else {
		log.Debugf("[Control] %s ban %d in all groups", c.Service, userID)
	}
}

// Permit permits plugin permisson for user in group
//
// if groupID is 0, it will permit the plugin for all groups
func (c *Control[Ctx]) Permit(groupID, userID int64) {
	var err error
	var digest [16]byte
	if groupID != 0 {
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", c.Service, userID, groupID)))
		id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
		c.Manager.Lock()
		err = c.Manager.DataBase.Del(c.Service+"ban", "WHERE id = "+strconv.FormatInt(int64(id), 10))
		delete(banCache, id)
		c.Manager.Unlock()
		if err != nil {
			log.Errorf("[Control] %s permit %d in group %d failed: %v", c.Service, userID, groupID, err)
		} else {
			log.Debugf("[Control] %s permit %d in group %d", c.Service, userID, groupID)
		}
		return
	}
	// permit all groups
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", c.Service, userID)))
	id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
	c.Manager.Lock()
	err = c.Manager.DataBase.Del(c.Service+"ban", "WHERE id = "+strconv.FormatInt(int64(id), 10))
	delete(banCache, id)
	c.Manager.Unlock()
	if err != nil {
		log.Errorf("[Control] %s permit %d in all groups failed: %v", c.Service, userID, err)
	} else {
		log.Debugf("[Control] %s permit %d in all groups", c.Service, userID)
	}
}

// IsBanned checks if the user is banned in group
func (c *Control[Ctx]) IsBanned(groupID, userID int64) bool {
	var digest [16]byte
	var err error
	var b BanStatus
	if groupID != 0 {
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_%d", c.Service, userID, groupID)))
		id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
		c.Manager.RLock()
		// read from cache
		if y, ok := banCache[id]; ok {
			c.Manager.RUnlock()
			return y
		}
		err := c.Manager.DataBase.Find(c.Service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(id), 10))
		c.Manager.RUnlock()
		if err == nil && b.UserID == userID && b.GroupID == groupID {
			log.Debugf("[Control] %s %d is banned in group %d", c.Service, userID, groupID)
			c.Manager.Lock()
			banCache[id] = true
			c.Manager.Unlock()
			return true
		}
		c.Manager.Lock()
		banCache[id] = false
		c.Manager.Unlock()
	}
	// check all groups
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("[%s]%d_all", c.Service, userID)))
	id := binary.LittleEndian.Uint64(digest[:8]) // 8 bytes as primary key
	c.Manager.RLock()
	// read from cache
	if y, ok := banCache[id]; ok {
		c.Manager.RUnlock()
		return y
	}
	err = c.Manager.DataBase.Find(c.Service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(id), 10))
	c.Manager.RUnlock()
	if err == nil && b.UserID == userID && b.GroupID == 0 {
		log.Debugf("[Control] %s %d is banned in all groups", c.Service, userID)
		c.Manager.Lock()
		banCache[id] = true
		c.Manager.Unlock()
		return true
	}
	c.Manager.Lock()
	banCache[id] = false
	c.Manager.Unlock()
	return false
}

// String returns the service help
func (c *Control[Ctx]) String() string {
	return c.Options.HelpText
}

// GetData return group setting data
func (c *Control[Ctx]) GetData(groupID int64) int64 {
	var gconf GroupConfig
	var err error
	c.Manager.RLock()
	err = c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	c.Manager.RUnlock()
	if err == nil && gconf.GroupID == groupID {
		log.Debugf("[Control] %s get data of group %d: %v", c.Service, groupID, gconf.Disable>>1)
		return (gconf.Disable >> 1) & 0x3fffffff_ffffffff
	}
	return 0
}

// SetData set group setting data
func (c *Control[Ctx]) SetData(groupID, data int64) error {
	var gconf GroupConfig
	c.Manager.RLock()
	err := c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	c.Manager.RUnlock()
	if err != nil {
		gconf.GroupID = groupID
		if c.Options.DefaultDisable {
			gconf.Disable = 1
		}
	}

	x := bits.RotateLeft64(uint64(gconf.Disable), 1)
	x &= 0x03
	x |= uint64(data) << 2
	gconf.Disable = int64(bits.RotateLeft64(x, -1))

	log.Debugf("[Control] %s set data of group %d: 0x%x", c.Service, groupID, data)
	c.Manager.Lock()
	err = c.Manager.DataBase.Insert(c.Service, &gconf)
	c.Manager.Unlock()
	if err != nil {
		log.Errorf("[Control] %s set data of group %d failed: %v", c.Service, groupID, err)
	}
	return err
}

// GetExtra return extra setting data
func (c *Control[Ctx]) GetExtra(obj any) error {
	if c.Options.Extra == 0 {
		return fmt.Errorf("GetExtra not implemented") // todo: add error
	}
	return c.Manager.getExtra(int64(c.Options.Extra), obj)
}

// SetExtra set extra setting data
func (c *Control[Ctx]) SetExtra(obj any) error {
	if c.Options.Extra == 0 {
		return fmt.Errorf("SetExtra not implemented") // todo: add error
	}
	_ = c.Manager.Response(int64(c.Options.Extra))
	return c.Manager.setExtra(int64(c.Options.Extra), obj)
}

// Flip flips default plugin status for all groups
func (c *Control[Ctx]) Flip() error {
	var gconf GroupConfig
	c.Manager.Lock()
	defer c.Manager.Unlock()
	c.Options.DefaultDisable = !c.Options.DefaultDisable
	err := c.Manager.DataBase.Find(c.Service, &gconf, "WHERE gid = 0")
	if err != nil && c.Options.DefaultDisable {
		gconf.Disable = 1
	}
	x := bits.RotateLeft64(uint64(gconf.Disable), 1) &^ 1
	gconf.Disable = int64(bits.RotateLeft64(x, -1))
	log.Debugf("[Control] flip plugin %s of all : %d %v", c.Service, gconf.GroupID, x&1)
	err = c.Manager.DataBase.Insert(c.Service, &gconf)
	if err != nil {
		log.Errorf("[Control] %s flip failed: %v", c.Service, err)
	}
	return err
}
