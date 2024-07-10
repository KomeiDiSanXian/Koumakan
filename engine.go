package zero

import (
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/FloatTech/floatbox/file"
	"github.com/sirupsen/logrus"
	"github.com/KomeiDiSanXian/Koumakan/extension/control"
)

// New 生成空引擎
func New() Engine {
	return &ZeroEngine{
		preHandler:  []Rule{},
		midHandler:  []Rule{},
		postHandler: []Handler{},
	}
}

var defaultEngine = New()

// ZeroEngine is the pre_handler, post_handler manager, it implements the Engine interface
type ZeroEngine struct {
	preHandler  []Rule
	midHandler  []Rule
	postHandler []Handler
	block       bool
	matchers    []IMatcher
	prio        int
	service     string
	datafolder  string
}

var prioMap = make(map[int]string)      // prioMap is map[prio]service
var briefMap = make(map[string]string)  // briefMap is map[brief]service
var folderMap = make(map[string]string) // folderMap is map[folder]service
var extraMap = make(map[int16]string)   // extraMap is map[gid]service

func newEngine(service string, prio int, o *control.Option[*Ctx]) Engine {
	eng := &ZeroEngine{
		prio:    prio,
		service: service,
	}
	s, ok := prioMap[prio]
	if ok {
		panic(fmt.Sprint("prio", prio, "is used by", s))
	}
	prioMap[prio] = service
	eng.UsePreHandler(
		func(ctx *Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
		newControl(service, o),
	)
	if o.Brief != "" {
		s, ok := briefMap[o.Brief]
		if ok {
			panic("Brief \"" + o.Brief + "\" of service " + service + " has been required by service " + s)
		}
		briefMap[o.Brief] = service
	}
	if o.Extra != 0 {
		s, ok := extraMap[o.Extra]
		if ok {
			panic("Extra " + strconv.Itoa(int(o.Extra)) + " of service " + service + " has been required by service " + s)
		}
		extraMap[o.Extra] = service
	}

	switch {
	case o.PublicDataFolder != "":
		if unicode.IsLower([]rune(o.PublicDataFolder)[0]) {
			panic("public data folder " + o.PublicDataFolder + " must start with an upper case letter")
		}
		eng.datafolder = "data/" + o.PublicDataFolder + "/"
	case o.PrivateDataFolder != "":
		if unicode.IsUpper([]rune(o.PrivateDataFolder)[0]) {
			panic("private data folder " + o.PrivateDataFolder + " must start with an lower case letter")
		}
	default:
		eng.datafolder = "data/zbp/"
	}

	if eng.datafolder != "data/zbp/" {
		s, ok := folderMap[eng.datafolder]
		if ok {
			panic("folder " + eng.datafolder + " has been required by service " + s)
		}
		folderMap[eng.datafolder] = service
	}
	if file.IsNotExist(eng.datafolder) {
		err := os.MkdirAll(eng.datafolder, 0755)
		if err != nil {
			panic(err)
		}
	}
	logrus.Debugf("[ZeroBot] Service %s has been loaded, data path: %s", service, eng.datafolder)
	return eng
}

func (e *ZeroEngine) getBlock() bool { return e.block }

func (e *ZeroEngine) getPreHandler() []Rule { return e.preHandler }

func (e *ZeroEngine) getMidHandler() []Rule { return e.midHandler }

func (e *ZeroEngine) getPostHandler() []Handler { return e.postHandler }

// Delete 移除该 ZeroEngine 注册的所有 Matchers
func (e *ZeroEngine) Delete() {
	for _, m := range e.matchers {
		m.Delete()
	}
}

// DataFolder 获取当前插件的数据文件夹
func (e *ZeroEngine) DataFolder() string { return e.datafolder }

// IsEnabled 获取当前插件是否启用
func (e *ZeroEngine) IsEnabled(id int64) bool {
	c, ok := managers.Lookup(e.service)
	if !ok {
		return false
	}
	return c.IsEnable(id)
}

func (e *ZeroEngine) SetBlock(block bool) Engine {
	e.block = block
	return e
}

// UsePreHandler 向该 ZeroEngine 添加新 PreHandler(Rule),
// 会在 Rule 判断前触发，如果 preHandler
// 没有通过，则 Rule, Matcher 不会触发
//
// 可用于分群组管理插件等
func (e *ZeroEngine) UsePreHandler(rules ...Rule) {
	e.preHandler = append(e.preHandler, rules...)
}

// UseMidHandler 向该 ZeroEngine 添加新 MidHandler(Rule),
// 会在 Rule 判断后， Matcher 触发前触发，如果 midHandler
// 没有通过，则 Matcher 不会触发
//
// 可用于速率限制等
func (e *ZeroEngine) UseMidHandler(rules ...Rule) {
	e.midHandler = append(e.midHandler, rules...)
}

// UsePostHandler 向该 ZeroEngine 添加新 PostHandler(Rule),
// 会在 Matcher 触发后触发，如果 PostHandler 返回 false,
// 则后续的 post handler 不会触发
//
// 可用于反并发等
func (e *ZeroEngine) UsePostHandler(handler ...Handler) {
	e.postHandler = append(e.postHandler, handler...)
}

// On 添加新的指定消息类型的匹配器(默认ZeroEngine)
func On(typ string, rules ...Rule) IMatcher { return defaultEngine.On(typ, rules...) }

// On 添加新的指定消息类型的匹配器
func (e *ZeroEngine) On(typ string, rules ...Rule) IMatcher {
	matcher := &Matcher{
		Type:   Type(typ),
		Rules:  rules,
		Engine: e,
	}
	e.matchers = append(e.matchers, matcher)
	return StoreMatcher(matcher)
}

// OnMessage 消息触发器
func OnMessage(rules ...Rule) IMatcher { return On("message", rules...) }

// OnMessage 消息触发器
func (e *ZeroEngine) OnMessage(rules ...Rule) IMatcher { return e.On("message", rules...) }

// OnNotice 系统提示触发器
func OnNotice(rules ...Rule) IMatcher { return On("notice", rules...) }

// OnNotice 系统提示触发器
func (e *ZeroEngine) OnNotice(rules ...Rule) IMatcher { return e.On("notice", rules...) }

// OnRequest 请求消息触发器
func OnRequest(rules ...Rule) IMatcher { return On("request", rules...) }

// OnRequest 请求消息触发器
func (e *ZeroEngine) OnRequest(rules ...Rule) IMatcher { return On("request", rules...) }

// OnMetaEvent 元事件触发器
func OnMetaEvent(rules ...Rule) IMatcher { return On("meta_event", rules...) }

// OnMetaEvent 元事件触发器
func (e *ZeroEngine) OnMetaEvent(rules ...Rule) IMatcher { return On("meta_event", rules...) }

// OnPrefix 前缀触发器
func OnPrefix(prefix string, rules ...Rule) IMatcher {
	return defaultEngine.OnPrefix(prefix, rules...)
}

// OnPrefix 前缀触发器
func (e *ZeroEngine) OnPrefix(prefix string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{PrefixRule(prefix)}, rules...)...)
}

// OnSuffix 后缀触发器
func OnSuffix(suffix string, rules ...Rule) IMatcher {
	return defaultEngine.OnSuffix(suffix, rules...)
}

// OnSuffix 后缀触发器
func (e *ZeroEngine) OnSuffix(suffix string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{SuffixRule(suffix)}, rules...)...)
}

// OnCommand 命令触发器
func OnCommand(commands string, rules ...Rule) IMatcher {
	return defaultEngine.OnCommand(commands, rules...)
}

// OnCommand 命令触发器
func (e *ZeroEngine) OnCommand(commands string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{CommandRule(commands)}, rules...)...)
}

// OnRegex 正则触发器
func OnRegex(regexPattern string, rules ...Rule) IMatcher {
	return OnMessage(append([]Rule{RegexRule(regexPattern)}, rules...)...)
}

// OnRegex 正则触发器
func (e *ZeroEngine) OnRegex(regexPattern string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{RegexRule(regexPattern)}, rules...)...)
}

// OnKeyword 关键词触发器
func OnKeyword(keyword string, rules ...Rule) IMatcher {
	return defaultEngine.OnKeyword(keyword, rules...)
}

// OnKeyword 关键词触发器
func (e *ZeroEngine) OnKeyword(keyword string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{KeywordRule(keyword)}, rules...)...)
}

// OnFullMatch 完全匹配触发器
func OnFullMatch(src string, rules ...Rule) IMatcher {
	return defaultEngine.OnFullMatch(src, rules...)
}

// OnFullMatch 完全匹配触发器
func (e *ZeroEngine) OnFullMatch(src string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{FullMatchRule(src)}, rules...)...)
}

// OnFullMatchGroup 完全匹配触发器组
func OnFullMatchGroup(src []string, rules ...Rule) IMatcher {
	return defaultEngine.OnFullMatchGroup(src, rules...)
}

// OnFullMatchGroup 完全匹配触发器组
func (e *ZeroEngine) OnFullMatchGroup(src []string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{FullMatchRule(src...)}, rules...)...)
}

// OnKeywordGroup 关键词触发器组
func OnKeywordGroup(keywords []string, rules ...Rule) IMatcher {
	return defaultEngine.OnKeywordGroup(keywords, rules...)
}

// OnKeywordGroup 关键词触发器组
func (e *ZeroEngine) OnKeywordGroup(keywords []string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{KeywordRule(keywords...)}, rules...)...)
}

// OnCommandGroup 命令触发器组
func OnCommandGroup(commands []string, rules ...Rule) IMatcher {
	return defaultEngine.OnCommandGroup(commands, rules...)
}

// OnCommandGroup 命令触发器组
func (e *ZeroEngine) OnCommandGroup(commands []string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{CommandRule(commands...)}, rules...)...)
}

// OnPrefixGroup 前缀触发器组
func OnPrefixGroup(prefix []string, rules ...Rule) IMatcher {
	return defaultEngine.OnPrefixGroup(prefix, rules...)
}

// OnPrefixGroup 前缀触发器组
func (e *ZeroEngine) OnPrefixGroup(prefix []string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{PrefixRule(prefix...)}, rules...)...)
}

// OnSuffixGroup 后缀触发器组
func OnSuffixGroup(suffix []string, rules ...Rule) IMatcher {
	return defaultEngine.OnSuffixGroup(suffix, rules...)
}

// OnSuffixGroup 后缀触发器组
func (e *ZeroEngine) OnSuffixGroup(suffix []string, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{SuffixRule(suffix...)}, rules...)...)
}

// OnShell shell命令触发器
func OnShell(command string, model any, rules ...Rule) IMatcher {
	return defaultEngine.OnShell(command, model, rules...)
}

// OnShell shell命令触发器
func (e *ZeroEngine) OnShell(command string, model any, rules ...Rule) IMatcher {
	return e.On("message", append([]Rule{ShellRule(command, model)}, rules...)...)
}
