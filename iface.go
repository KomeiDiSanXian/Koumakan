package zero

import (
	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// IEngineTrigger 是 ZeroBot 触发器的接口
type IEngineTrigger interface {
	On(typ string, rules ...Rule) IMatcher
	OnCommand(commands string, rules ...Rule) IMatcher
	OnCommandGroup(commands []string, rules ...Rule) IMatcher
	OnFullMatch(src string, rules ...Rule) IMatcher
	OnFullMatchGroup(src []string, rules ...Rule) IMatcher
	OnKeyword(keyword string, rules ...Rule) IMatcher
	OnKeywordGroup(keywords []string, rules ...Rule) IMatcher
	OnMessage(rules ...Rule) IMatcher
	OnMetaEvent(rules ...Rule) IMatcher
	OnNotice(rules ...Rule) IMatcher
	OnPrefix(prefix string, rules ...Rule) IMatcher
	OnPrefixGroup(prefix []string, rules ...Rule) IMatcher
	OnRegex(regexPattern string, rules ...Rule) IMatcher
	OnRequest(rules ...Rule) IMatcher
	OnShell(command string, model interface{}, rules ...Rule) IMatcher
	OnSuffix(suffix string, rules ...Rule) IMatcher
	OnSuffixGroup(suffix []string, rules ...Rule) IMatcher
}

// IEngineHandler 是 ZeroBot 处理器的接口
type IEngineHandler interface {
	UseMidHandler(rules ...Rule)
	UsePostHandler(handler ...Handler)
	UsePreHandler(rules ...Rule)
}

// IEngine 是 ZeroBot 引擎的接口
type IEngine interface {
	IEngineTrigger
	IEngineHandler
	Delete()
	SetBlock(block bool) IEngine

	getPreHandler() []Rule
	getMidHandler() []Rule
	getPostHandler() []Handler
	getBlock() bool
}

// IMatcherSetter ...
type IMatcherSetter interface {
	BindEngine(e IEngine) IMatcher
	FirstPriority() IMatcher
	SecondPriority() IMatcher
	SetBlock(block bool) IMatcher
	SetPriority(priority int) IMatcher
	ThirdPriority() IMatcher
	SetBreak(b bool) IMatcher
	SetNoTimeout(b bool) IMatcher
	SetRules(rules ...Rule) IMatcher
}

// IMatcherGetter ...
type IMatcherGetter interface {
	GetPriority() int
	GetBlock() bool
	GetTemp() bool
	GetNoTimeout() bool
	GetBreak() bool
	GetType() Rule
	GetRules() []Rule
	GetHandler() Handler
	GetEngine() IEngine
}

// IMatcher 是 ZeroBot 匹配器的接口
type IMatcher interface {
	IMatcherSetter
	IMatcherGetter
	Delete()
	FutureEvent(Type string, rule ...Rule) *FutureEvent
	Handle(handler Handler) IMatcher
	Limit(limiterfn func(Context) *rate.Limiter, postfn ...func(Context)) IMatcher
}

// ContextGetter ...
type ContextGetter interface {
	GetEvent() *Event
	GetState() State
}

// Context 上下文接口
type Context interface {
	OneBotAPI
	GoCQAPI
	LLoneBotAPI
	ContextGetter
	Block()
	Break()
	CallAction(action string, params Params) APIResponse
	CardOrNickName(uid int64) string
	CheckSession() Rule
	Echo(response []byte)
	ExtractPlainText() string
	FutureEvent(Type string, rule ...Rule) *FutureEvent
	Get(prompt string) string
	GetMatcher() IMatcher
	MessageString() string
	NickName() string
	NoTimeout()
	Parse(model interface{}) error
	Send(msg interface{}) message.MessageID
	SendChain(msg ...message.MessageSegment) message.MessageID

	setMatcher(matcher IMatcher)
	getMatcher() IMatcher
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
