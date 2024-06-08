package zero

// New 生成空引擎
func New() Engine {
	return &ZeroEngine{
		preHandler:  []Rule{},
		midHandler:  []Rule{},
		postHandler: []Handler{},
	}
}

// EngineBase 基础接口
type EngineBase interface {
	Delete()                           // 删除该 Engine 注册的所有 Matchers
	SetBlock(block bool) Engine        // 设置是否阻断后续 Matcher 触发
	UsePreHandler(rules ...Rule)       // 添加新 PreHandler(Rule)
	UseMidHandler(rules ...Rule)       // 添加新 MidHandler(Rule)
	UsePostHandler(handler ...Handler) // 添加新 PostHandler(Rule)

}

type getter interface {
	getBlock() bool
	getPreHandler() []Rule
	getMidHandler() []Rule
	getPostHandler() []Handler
}

// Engine 引擎接口
type Engine interface {
	getter
	EngineBase
	EngineTrigger
	EngineMessage
}

// EngineTrigger 触发器接口
type EngineTrigger interface {
	OnPrefix(prefix string, rules ...Rule) *Matcher            // 前缀触发器
	OnSuffix(suffix string, rules ...Rule) *Matcher            // 后缀触发器
	OnCommand(commands string, rules ...Rule) *Matcher         // 命令触发器
	OnRegex(regexPattern string, rules ...Rule) *Matcher       // 正则触发器
	OnKeyword(keyword string, rules ...Rule) *Matcher          // 关键词触发器
	OnFullMatch(src string, rules ...Rule) *Matcher            // 完全匹配触发器
	OnFullMatchGroup(src []string, rules ...Rule) *Matcher     // 完全匹配触发器组
	OnKeywordGroup(keywords []string, rules ...Rule) *Matcher  // 关键词触发器组
	OnCommandGroup(commands []string, rules ...Rule) *Matcher  // 命令触发器组
	OnPrefixGroup(prefix []string, rules ...Rule) *Matcher     // 前缀触发器组
	OnSuffixGroup(suffix []string, rules ...Rule) *Matcher     // 后缀触发器组
	OnShell(command string, model any, rules ...Rule) *Matcher // shell命令触发器
}

// EngineMessage 消息触发器接口
type EngineMessage interface {
	On(typ string, rules ...Rule) *Matcher // 添加新的指定消息类型的匹配器
	OnMessage(rules ...Rule) *Matcher      // 消息触发器
	OnNotice(rules ...Rule) *Matcher       // 系统提示触发器
	OnRequest(rules ...Rule) *Matcher      // 请求消息触发器
	OnMetaEvent(rules ...Rule) *Matcher    // 元事件触发器
}

var defaultEngine = New()

// ZeroEngine is the pre_handler, post_handler manager, it implements the Engine interface
type ZeroEngine struct {
	preHandler  []Rule
	midHandler  []Rule
	postHandler []Handler
	block       bool
	matchers    []*Matcher
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
func On(typ string, rules ...Rule) *Matcher { return defaultEngine.On(typ, rules...) }

// On 添加新的指定消息类型的匹配器
func (e *ZeroEngine) On(typ string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type(typ),
		Rules:  rules,
		Engine: e,
	}
	e.matchers = append(e.matchers, matcher)
	return StoreMatcher(matcher)
}

// OnMessage 消息触发器
func OnMessage(rules ...Rule) *Matcher { return On("message", rules...) }

// OnMessage 消息触发器
func (e *ZeroEngine) OnMessage(rules ...Rule) *Matcher { return e.On("message", rules...) }

// OnNotice 系统提示触发器
func OnNotice(rules ...Rule) *Matcher { return On("notice", rules...) }

// OnNotice 系统提示触发器
func (e *ZeroEngine) OnNotice(rules ...Rule) *Matcher { return e.On("notice", rules...) }

// OnRequest 请求消息触发器
func OnRequest(rules ...Rule) *Matcher { return On("request", rules...) }

// OnRequest 请求消息触发器
func (e *ZeroEngine) OnRequest(rules ...Rule) *Matcher { return On("request", rules...) }

// OnMetaEvent 元事件触发器
func OnMetaEvent(rules ...Rule) *Matcher { return On("meta_event", rules...) }

// OnMetaEvent 元事件触发器
func (e *ZeroEngine) OnMetaEvent(rules ...Rule) *Matcher { return On("meta_event", rules...) }

// OnPrefix 前缀触发器
func OnPrefix(prefix string, rules ...Rule) *Matcher {
	return defaultEngine.OnPrefix(prefix, rules...)
}

// OnPrefix 前缀触发器
func (e *ZeroEngine) OnPrefix(prefix string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{PrefixRule(prefix)}, rules...)...)
}

// OnSuffix 后缀触发器
func OnSuffix(suffix string, rules ...Rule) *Matcher {
	return defaultEngine.OnSuffix(suffix, rules...)
}

// OnSuffix 后缀触发器
func (e *ZeroEngine) OnSuffix(suffix string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{SuffixRule(suffix)}, rules...)...)
}

// OnCommand 命令触发器
func OnCommand(commands string, rules ...Rule) *Matcher {
	return defaultEngine.OnCommand(commands, rules...)
}

// OnCommand 命令触发器
func (e *ZeroEngine) OnCommand(commands string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{CommandRule(commands)}, rules...)...)
}

// OnRegex 正则触发器
func OnRegex(regexPattern string, rules ...Rule) *Matcher {
	return OnMessage(append([]Rule{RegexRule(regexPattern)}, rules...)...)
}

// OnRegex 正则触发器
func (e *ZeroEngine) OnRegex(regexPattern string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{RegexRule(regexPattern)}, rules...)...)
}

// OnKeyword 关键词触发器
func OnKeyword(keyword string, rules ...Rule) *Matcher {
	return defaultEngine.OnKeyword(keyword, rules...)
}

// OnKeyword 关键词触发器
func (e *ZeroEngine) OnKeyword(keyword string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{KeywordRule(keyword)}, rules...)...)
}

// OnFullMatch 完全匹配触发器
func OnFullMatch(src string, rules ...Rule) *Matcher {
	return defaultEngine.OnFullMatch(src, rules...)
}

// OnFullMatch 完全匹配触发器
func (e *ZeroEngine) OnFullMatch(src string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{FullMatchRule(src)}, rules...)...)
}

// OnFullMatchGroup 完全匹配触发器组
func OnFullMatchGroup(src []string, rules ...Rule) *Matcher {
	return defaultEngine.OnFullMatchGroup(src, rules...)
}

// OnFullMatchGroup 完全匹配触发器组
func (e *ZeroEngine) OnFullMatchGroup(src []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{FullMatchRule(src...)}, rules...)...)
}

// OnKeywordGroup 关键词触发器组
func OnKeywordGroup(keywords []string, rules ...Rule) *Matcher {
	return defaultEngine.OnKeywordGroup(keywords, rules...)
}

// OnKeywordGroup 关键词触发器组
func (e *ZeroEngine) OnKeywordGroup(keywords []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{KeywordRule(keywords...)}, rules...)...)
}

// OnCommandGroup 命令触发器组
func OnCommandGroup(commands []string, rules ...Rule) *Matcher {
	return defaultEngine.OnCommandGroup(commands, rules...)
}

// OnCommandGroup 命令触发器组
func (e *ZeroEngine) OnCommandGroup(commands []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{CommandRule(commands...)}, rules...)...)
}

// OnPrefixGroup 前缀触发器组
func OnPrefixGroup(prefix []string, rules ...Rule) *Matcher {
	return defaultEngine.OnPrefixGroup(prefix, rules...)
}

// OnPrefixGroup 前缀触发器组
func (e *ZeroEngine) OnPrefixGroup(prefix []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{PrefixRule(prefix...)}, rules...)...)
}

// OnSuffixGroup 后缀触发器组
func OnSuffixGroup(suffix []string, rules ...Rule) *Matcher {
	return defaultEngine.OnSuffixGroup(suffix, rules...)
}

// OnSuffixGroup 后缀触发器组
func (e *ZeroEngine) OnSuffixGroup(suffix []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{SuffixRule(suffix...)}, rules...)...)
}

// OnShell shell命令触发器
func OnShell(command string, model any, rules ...Rule) *Matcher {
	return defaultEngine.OnShell(command, model, rules...)
}

// OnShell shell命令触发器
func (e *ZeroEngine) OnShell(command string, model any, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{ShellRule(command, model)}, rules...)...)
}
