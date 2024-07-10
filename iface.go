package zero

import (
	"github.com/KomeiDiSanXian/Koumakan/message"
	"github.com/tidwall/gjson"
)

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

// Lazy 懒加载接口
type Lazy interface {
	GetLazyData(filename string, isDataMustEqual bool) ([]byte, error)
	GetCustomLazyData(dataurl, filename string) ([]byte, error)
	InitWhenNoError(errfun func() error, do func())
}

// Engine 引擎接口
type Engine interface {
	getter
	EngineBase
	EngineTrigger
	EngineMessage
	DataFolder() string      // 获取当前插件的数据文件夹
	IsEnabled(id int64) bool // 获取当前插件是否启用

	Lazy
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

// Context 上下文接口
type Context interface {
	OneBotAPI
	GoCQAPI
	LLoneBotAPI
}

// OneBotAPI OneBotAPI接口
type OneBotAPI interface {
	SendGroupMessage(groupID int64, message any) int64
	SendPrivateMessage(userID int64, message any) int64
	DeleteMessage(messageID any)
	GetMessage(messageID any) Message
	GetForwardMessage(id string) gjson.Result
	SendLike(userID int64, times int)
	SetGroupKick(groupID, userID int64, rejectAddRequest bool)
	SetThisGroupKick(userID int64, rejectAddRequest bool)
	SetGroupBan(groupID, userID, duration int64)
	SetThisGroupBan(userID, duration int64)
	SetGroupWholeBan(groupID int64, enable bool)
	SetThisGroupWholeBan(enable bool)
	SetGroupAdmin(groupID, userID int64, enable bool)
	SetThisGroupAdmin(userID int64, enable bool)
	SetGroupAnonymous(groupID int64, enable bool)
	SetThisGroupAnonymous(enable bool)
	SetGroupCard(groupID, userID int64, card string)
	SetThisGroupCard(userID int64, card string)
	SetGroupName(groupID int64, groupName string)
	SetThisGroupName(groupID int64, groupName string)
	SetGroupLeave(groupID int64, isDismiss bool)
	SetThisGroupLeave(isDismiss bool)
	SetGroupSpecialTitle(groupID, userID int64, specialTitle string)
	SetThisGroupSpecialTitle(userID int64, specialTitle string)
	SetFriendAddRequest(flag string, approve bool, remark string)
	SetGroupAddRequest(flag string, subType string, approve bool, reason string)
	GetLoginInfo() gjson.Result
	GetStrangerInfo(userID int64, noCache bool) gjson.Result
	GetFriendList() gjson.Result
	GetGroupInfo(groupID int64, noCache bool) Group
	GetThisGroupInfo(noCache bool) Group
	GetGroupList() gjson.Result
	GetGroupMemberInfo(groupID int64, userID int64, noCache bool) gjson.Result
	GetThisGroupMemberInfo(userID int64, noCache bool) gjson.Result
	GetGroupMemberList(groupID int64) gjson.Result
	GetThisGroupMemberList() gjson.Result
	GetGroupMemberListNoCache(groupID int64) gjson.Result
	GetThisGroupMemberListNoCache() gjson.Result
	GetGroupHonorInfo(groupID int64, hType string) gjson.Result
	GetThisGroupHonorInfo(hType string) gjson.Result
	GetRecord(file string, outFormat string) gjson.Result
	GetImage(file string) gjson.Result
	GetVersionInfo() gjson.Result
}

// GoCQAPI GoCQAPI接口
type GoCQAPI interface {
	SetGroupPortrait(groupID int64, file string)
	SetThisGroupPortrait(file string)
	OCRImage(file string) gjson.Result
	SendGroupForwardMessage(groupID int64, message message.Message) gjson.Result
	SendPrivateForwardMessage(userID int64, message message.Message) gjson.Result
	GetGroupSystemMessage() gjson.Result
	MarkMessageAsRead(messageID int64) APIResponse
	MarkThisMessageAsRead() APIResponse
	GetOnlineClients(noCache bool) gjson.Result
	GetGroupAtAllRemain(groupID int64) gjson.Result
	GetThisGroupAtAllRemain() gjson.Result
	GetGroupMessageHistory(groupID, messageID int64) gjson.Result
	GetLatestGroupMessageHistory(groupID int64) gjson.Result
	GetThisGroupMessageHistory(messageID int64) gjson.Result
	GetLatestThisGroupMessageHistory() gjson.Result
	GetGroupEssenceMessageList(groupID int64) gjson.Result
	GetThisGroupEssenceMessageList() gjson.Result
	SetGroupEssenceMessage(messageID int64) APIResponse
	DeleteGroupEssenceMessage(messageID int64) APIResponse
	GetWordSlices(content string) gjson.Result
	GetGroupFilesystemInfo(groupID int64) gjson.Result
	GetThisGroupFilesystemInfo() gjson.Result
	GetGroupRootFiles(groupID int64) gjson.Result
	GetThisGroupRootFiles() gjson.Result
	GetGroupFilesByFolder(groupID int64, folderID string) gjson.Result
	GetThisGroupFilesByFolder(folderID string) gjson.Result
	GetGroupFileUrl(groupID, busid int64, fileID string) string
	GetThisGroupFileUrl(busid int64, fileID string) string
	UploadGroupFile(groupID int64, file, name, folder string) APIResponse
	UploadThisGroupFile(file, name, folder string) APIResponse
}

// llOneBotAPI llOneBotAPI接口
type LLoneBotAPI interface {
	ForwardFriendSingleMessage(userID int64, messageID any) APIResponse
	ForwardGroupSingleMessage(groupID int64, messageID any) APIResponse
	SetMyAvatar(file string) APIResponse
	GetFile(fileID string) gjson.Result
	SetMessageEmojiLike(messageID any, emojiID rune) error
}
