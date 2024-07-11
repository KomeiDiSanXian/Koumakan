package zero

import (
	"hash/crc64"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

// Type check the ctx.Event's type
func Type(type_ string) Rule {
	t := strings.SplitN(type_, "/", 3)
	return func(ctx Context) bool {
		if len(t) > 0 && t[0] != ctx.GetEvent().PostType {
			return false
		}
		if len(t) > 1 && t[1] != ctx.GetEvent().DetailType {
			return false
		}
		if len(t) > 2 && t[2] != ctx.GetEvent().SubType {
			return false
		}
		return true
	}
}

// PrefixRule check if the message has the prefix and trim the prefix
//
// 检查消息前缀
func PrefixRule(prefixes ...string) Rule {
	return func(ctx Context) bool {
		if len(ctx.GetEvent().Message) == 0 || ctx.GetEvent().Message[0].Type != "text" { // 确保无空指针
			return false
		}
		first := ctx.GetEvent().Message[0]
		firstMessage := first.Data["text"]
		for _, prefix := range prefixes {
			if strings.HasPrefix(firstMessage, prefix) {
				ctx.GetState()["prefix"] = prefix
				arg := strings.TrimLeft(firstMessage[len(prefix):], " ")
				if len(ctx.GetEvent().Message) > 1 {
					arg += ctx.GetEvent().Message[1:].ExtractPlainText()
				}
				ctx.GetState()["args"] = arg
				return true
			}
		}
		return false
	}
}

// SuffixRule check if the message has the suffix and trim the suffix
//
// 检查消息后缀
func SuffixRule(suffixes ...string) Rule {
	return func(ctx Context) bool {
		mLen := len(ctx.GetEvent().Message)
		if mLen <= 0 { // 确保无空指针
			return false
		}
		last := ctx.GetEvent().Message[mLen-1]
		if last.Type != "text" {
			return false
		}
		lastMessage := last.Data["text"]
		for _, suffix := range suffixes {
			if strings.HasSuffix(lastMessage, suffix) {
				ctx.GetState()["suffix"] = suffix
				arg := strings.TrimRight(lastMessage[:len(lastMessage)-len(suffix)], " ")
				if mLen >= 2 {
					arg += ctx.GetEvent().Message[:mLen].ExtractPlainText()
				}
				ctx.GetState()["args"] = arg
				return true
			}
		}
		return false
	}
}

// CommandRule check if the message is a command and trim the command name
func CommandRule(commands ...string) Rule {
	return func(ctx Context) bool {
		if len(ctx.GetEvent().Message) == 0 || ctx.GetEvent().Message[0].Type != "text" {
			return false
		}
		first := ctx.GetEvent().Message[0]
		firstMessage := first.Data["text"]
		if !strings.HasPrefix(firstMessage, BotConfig.CommandPrefix) {
			return false
		}
		cmdMessage := firstMessage[len(BotConfig.CommandPrefix):]
		for _, command := range commands {
			if strings.HasPrefix(cmdMessage, command) {
				ctx.GetState()["command"] = command
				arg := strings.TrimLeft(cmdMessage[len(command):], " ")
				if len(ctx.GetEvent().Message) > 1 {
					arg += ctx.GetEvent().Message[1:].ExtractPlainText()
				}
				ctx.GetState()["args"] = arg
				return true
			}
		}
		return false
	}
}

// RegexRule check if the message can be matched by the regex pattern
func RegexRule(regexPattern string) Rule {
	regex := regexp.MustCompile(regexPattern)
	return func(ctx Context) bool {
		msg := ctx.MessageString()
		if matched := regex.FindStringSubmatch(msg); matched != nil {
			ctx.GetState()["regex_matched"] = matched
			return true
		}
		return false
	}
}

// ReplyRule check if the message is replying some message
func ReplyRule(messageID int64) Rule {
	return func(ctx Context) bool {
		if len(ctx.GetEvent().Message) == 0 {
			return false
		}
		if ctx.GetEvent().Message[0].Type != "reply" {
			return false
		}
		if id, err := strconv.ParseInt(ctx.GetEvent().Message[0].Data["id"], 10, 64); err == nil {
			return id == messageID
		}
		c := crc64.New(crc64.MakeTable(crc64.ISO))
		c.Write(helper.StringToBytes(ctx.GetEvent().Message[0].Data["id"]))
		return int64(c.Sum64()) == messageID
	}
}

// KeywordRule check if the message has a keyword or keywords
func KeywordRule(src ...string) Rule {
	return func(ctx Context) bool {
		msg := ctx.MessageString()
		for _, str := range src {
			if strings.Contains(msg, str) {
				ctx.GetState()["keyword"] = str
				return true
			}
		}
		return false
	}
}

// FullMatchRule check if src has the same copy of the message
func FullMatchRule(src ...string) Rule {
	return func(ctx Context) bool {
		msg := ctx.MessageString()
		for _, str := range src {
			if str == msg {
				ctx.GetState()["matched"] = msg
				return true
			}
		}
		return false
	}
}

// OnlyToMe only triggered in conditions of @bot or begin with the nicknames
func OnlyToMe(ctx Context) bool {
	return ctx.GetEvent().IsToMe
}

// CheckUser only triggered by specific person
func CheckUser(userId ...int64) Rule {
	return func(ctx Context) bool {
		for _, uid := range userId {
			if ctx.GetEvent().UserID == uid {
				return true
			}
		}
		return false
	}
}

// CheckGroup only triggered in specific group
func CheckGroup(grpId ...int64) Rule {
	return func(ctx Context) bool {
		for _, gid := range grpId {
			if ctx.GetEvent().GroupID == gid {
				return true
			}
		}
		return false
	}
}

// OnlyPrivate requires that the ctx.Event is private message
func OnlyPrivate(ctx Context) bool {
	return ctx.GetEvent().PostType == "message" && ctx.GetEvent().DetailType == "private"
}

// OnlyPublic requires that the ctx.Event is public/group or public/guild message
func OnlyPublic(ctx Context) bool {
	return ctx.GetEvent().PostType == "message" && (ctx.GetEvent().DetailType == "group" || ctx.GetEvent().DetailType == "guild")
}

// OnlyGroup requires that the ctx.Event is public/group message
func OnlyGroup(ctx Context) bool {
	return ctx.GetEvent().PostType == "message" && ctx.GetEvent().DetailType == "group"
}

// OnlyGuild requires that the ctx.Event is public/guild message
func OnlyGuild(ctx Context) bool {
	return ctx.GetEvent().PostType == "message" && ctx.GetEvent().DetailType == "guild"
}

func issu(id int64) bool {
	for _, su := range BotConfig.SuperUsers {
		if su == id {
			return true
		}
	}
	return false
}

// SuperUserPermission only triggered by the bot's owner
func SuperUserPermission(ctx Context) bool {
	return issu(ctx.GetEvent().UserID)
}

// AdminPermission only triggered by the group admins or higher permission
func AdminPermission(ctx Context) bool {
	return SuperUserPermission(ctx) || ctx.GetEvent().Sender.Role == "owner" || ctx.GetEvent().Sender.Role == "admin"
}

// OwnerPermission only triggered by the group owner or higher permission
func OwnerPermission(ctx Context) bool {
	return SuperUserPermission(ctx) || ctx.GetEvent().Sender.Role == "owner"
}

// UserOrGrpAdmin 允许用户单独使用或群管使用
func UserOrGrpAdmin(ctx Context) bool {
	if OnlyGroup(ctx) {
		return AdminPermission(ctx)
	}
	return OnlyToMe(ctx)
}

// GroupHigherPermission 群发送者权限高于 target
//
// 隐含 OnlyGroup 判断
func GroupHigherPermission(gettarget func(ctx Context) int64) Rule {
	return func(ctx Context) bool {
		if !OnlyGroup(ctx) {
			return false
		}
		target := gettarget(ctx)
		if target == ctx.GetEvent().UserID { // 特判, 自己和自己比
			return false
		}
		if SuperUserPermission(ctx) {
			sender := ctx.GetEvent().UserID
			return BotConfig.GetFirstSuperUser(sender, target) == sender
		}
		if ctx.GetEvent().Sender.Role == "owner" {
			return !issu(target) && ctx.GetThisGroupMemberInfo(target, false).Get("role").Str != "owner"
		}
		if ctx.GetEvent().Sender.Role == "admin" {
			tgtrole := ctx.GetThisGroupMemberInfo(target, false).Get("role").Str
			return !issu(target) && tgtrole != "owner" && tgtrole != "admin"
		}
		return false // member is the lowest
	}
}

// HasPicture 消息含有图片返回 true
func HasPicture(ctx Context) bool {
	urls := []string{}
	for _, elem := range ctx.GetEvent().Message {
		if elem.Type == "image" {
			if elem.Data["url"] != "" {
				urls = append(urls, elem.Data["url"])
			}
		}
	}
	if len(urls) > 0 {
		ctx.GetState()["image_url"] = urls
		return true
	}
	return false
}

// MustProvidePicture 消息不存在图片阻塞120秒至有图片，超时返回 false
func MustProvidePicture(ctx Context) bool {
	if HasPicture(ctx) {
		return true
	}
	// 没有图片就索取
	ctx.SendChain(message.Text("请发送一张图片"))
	next := NewFutureEvent("message", 999, true, ctx.CheckSession(), HasPicture).Next()
	select {
	case <-time.After(time.Second * 120):
		return false
	case newctx := <-next:
		ctx.GetState()["image_url"] = newctx.GetState()["image_url"]
		ctx.GetEvent().MessageID = newctx.GetEvent().MessageID
		return true
	}
}
