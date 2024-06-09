package zero

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
	OnPrefix(prefix string, rules ...Rule) IMatcher            // 前缀触发器
	OnSuffix(suffix string, rules ...Rule) IMatcher            // 后缀触发器
	OnCommand(commands string, rules ...Rule) IMatcher         // 命令触发器
	OnRegex(regexPattern string, rules ...Rule) IMatcher       // 正则触发器
	OnKeyword(keyword string, rules ...Rule) IMatcher          // 关键词触发器
	OnFullMatch(src string, rules ...Rule) IMatcher            // 完全匹配触发器
	OnFullMatchGroup(src []string, rules ...Rule) IMatcher     // 完全匹配触发器组
	OnKeywordGroup(keywords []string, rules ...Rule) IMatcher  // 关键词触发器组
	OnCommandGroup(commands []string, rules ...Rule) IMatcher  // 命令触发器组
	OnPrefixGroup(prefix []string, rules ...Rule) IMatcher     // 前缀触发器组
	OnSuffixGroup(suffix []string, rules ...Rule) IMatcher     // 后缀触发器组
	OnShell(command string, model any, rules ...Rule) IMatcher // shell命令触发器
}

// EngineMessage 消息触发器接口
type EngineMessage interface {
	On(typ string, rules ...Rule) IMatcher // 添加新的指定消息类型的匹配器
	OnMessage(rules ...Rule) IMatcher      // 消息触发器
	OnNotice(rules ...Rule) IMatcher       // 系统提示触发器
	OnRequest(rules ...Rule) IMatcher      // 请求消息触发器
	OnMetaEvent(rules ...Rule) IMatcher    // 元事件触发器
}

// IMatcher 匹配器接口
type IMatcher interface {
	MatcherGetter
	SetBlock(block bool) IMatcher                       // 设置是否阻断后续 Matcher 触发
	SetPriority(priority int) IMatcher                  // 设置当前 Matcher 优先级
	BindEngine(e Engine) IMatcher                       // 绑定当前 Matcher 到指定 Engine
	Delete()                                            // 删除当前 Matcher
	Handle(handler Handler) IMatcher                    // 设置当前 Matcher 的处理函数
	FutureEvent(Type string, rule ...Rule) *FutureEvent // 设置当前 Matcher 的 FutureEvent

	copy() IMatcher
}

// MatcherGetter 匹配器获取器接口
type MatcherGetter interface {
	GetPriority() int    // 获取当前 Matcher 优先级
	GetType() Rule       // 获取当前 Matcher 匹配的事件类型
	IsBlock() bool       // 是否阻断后续 Matcher
	IsBreak() bool       // 是否退出后续匹配流程
	IsNoTimeout() bool   // 是否不设超时
	IsTemp() bool        // 是否为临时Matcher
	GetRules() []Rule    // 获取当前 Matcher 的匹配规则
	GetHandler() Handler // 获取当前 Matcher 的处理函数
	GetEngine() Engine   // 获取当前 Matcher 的 Engine
}
