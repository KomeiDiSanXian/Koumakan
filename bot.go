package zero

import (
	"encoding/json"
	"hash/crc64"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FloatTech/ttl"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

// Config is config of zero bot
type Config struct {
	NickName       []string      `json:"nickname"`         // 机器人名称
	CommandPrefix  string        `json:"command_prefix"`   // 触发命令
	SuperUsers     []int64       `json:"super_users"`      // 超级用户
	RingLen        uint          `json:"ring_len"`         // 事件环长度 (默认关闭)
	Latency        time.Duration `json:"latency"`          // 事件处理延迟 (延迟 latency 再处理事件，在 ring 模式下不可低于 1ms)
	MaxProcessTime time.Duration `json:"max_process_time"` // 事件最大处理时间 (默认4min)
	MarkMessage    bool          `json:"mark_message"`     // 自动标记消息为已读
	Driver         []Driver      `json:"-"`                // 通信驱动
}

// APICallers 所有的APICaller列表， 通过self-ID映射
var APICallers callerMap

// APICaller is the interface of CallApi
type APICaller interface {
	CallApi(request APIRequest) (APIResponse, error)
}

// Driver 与OneBot通信的驱动，使用driver.DefaultWebSocketDriver
type Driver interface {
	Connect()
	Listen(func([]byte, APICaller))
}

// BotConfig 运行中bot的配置，是Run函数的参数的拷贝
var BotConfig Config

var (
	evring    eventRing // evring 事件环
	isrunning uintptr
)

func runinit(op *Config) {
	if op.MaxProcessTime == 0 {
		op.MaxProcessTime = time.Minute * 4
	}
	BotConfig = *op
	if op.RingLen == 0 {
		return
	}
	evring = newring(op.RingLen)
	evring.loop(op.Latency, op.MaxProcessTime, processEventAsync)
}

func (op *Config) directlink(b []byte, c APICaller) {
	go func() {
		if op.Latency != 0 {
			time.Sleep(op.Latency)
		}
		processEventAsync(b, c, op.MaxProcessTime)
	}()
}

// Run 主函数初始化
func Run(op *Config) {
	if !atomic.CompareAndSwapUintptr(&isrunning, 0, 1) {
		log.Warnln("[bot] 已忽略重复调用的 Run")
	}
	runinit(op)
	linkf := op.directlink
	if op.RingLen != 0 {
		linkf = evring.processEvent
	}
	for _, driver := range op.Driver {
		driver.Connect()
		go driver.Listen(linkf)
	}
}

// RunAndBlock 主函数初始化并阻塞
//
//	preblock 在所有 Driver 连接后，调用最后一个 Driver 的 Listen 阻塞前执行本函数
func RunAndBlock(op *Config, preblock func()) {
	if !atomic.CompareAndSwapUintptr(&isrunning, 0, 1) {
		log.Warnln("[bot] 已忽略重复调用的 RunAndBlock")
	}
	runinit(op)
	linkf := op.directlink
	if op.RingLen != 0 {
		linkf = evring.processEvent
	}
	switch len(op.Driver) {
	case 0:
		return
	case 1:
		op.Driver[0].Connect()
		if preblock != nil {
			preblock()
		}
		op.Driver[0].Listen(linkf)
	default:
		i := 0
		for ; i < len(op.Driver)-1; i++ {
			op.Driver[i].Connect()
			go op.Driver[i].Listen(linkf)
		}
		op.Driver[i].Connect()
		if preblock != nil {
			preblock()
		}
		op.Driver[i].Listen(linkf)
	}
}

var (
	triggeredMessages   = ttl.NewCache[int64, []message.MessageID](time.Minute * 5)
	triggeredMessagesMu = sync.Mutex{}
)

type messageLogger struct {
	msgid  message.MessageID
	caller APICaller
}

// CallApi 记录被触发的回复消息
func (m *messageLogger) CallApi(request APIRequest) (rsp APIResponse, err error) {
	rsp, err = m.caller.CallApi(request)
	if err != nil {
		return
	}
	id := rsp.Data.Get("message_id")
	if id.Exists() {
		mid := m.msgid.ID()
		triggeredMessagesMu.Lock()
		defer triggeredMessagesMu.Unlock()
		triggeredMessages.Set(mid,
			append(
				triggeredMessages.Get(mid),
				message.NewMessageIDFromString(id.String()),
			),
		)
	}
	return
}

// GetTriggeredMessages 获取被 id 消息触发的回复消息 id
func GetTriggeredMessages(id message.MessageID) []message.MessageID {
	triggeredMessagesMu.Lock()
	defer triggeredMessagesMu.Unlock()
	return triggeredMessages.Get(id.ID())
}

// processEventAsync 从池中处理事件, 异步调用匹配 mather
func processEventAsync(response []byte, caller APICaller, maxwait time.Duration) {
	var event Event
	_ = json.Unmarshal(response, &event)
	event.RawEvent = gjson.Parse(helper.BytesToString(response))
	var msgid message.MessageID
	messageID, err := strconv.ParseInt(helper.BytesToString(event.RawMessageID), 10, 64)
	if err == nil {
		event.MessageID = messageID
		msgid = message.NewMessageIDFromInteger(messageID)
	} else if event.MessageType == "guild" {
		// 是 guild 消息，进行如下转换以适配非 guild 插件
		// MessageID 填为 string
		event.MessageID, _ = strconv.Unquote(helper.BytesToString(event.RawMessageID))
		// 伪造 GroupID
		crc := crc64.New(crc64.MakeTable(crc64.ISO))
		crc.Write(helper.StringToBytes(event.GuildID))
		crc.Write(helper.StringToBytes(event.ChannelID))
		r := int64(crc.Sum64() & 0x7fff_ffff_ffff_ffff) // 确保为正数
		if r <= 0xffff_ffff {
			r |= 0x1_0000_0000 // 确保不与正常号码重叠
		}
		event.GroupID = r
		// 伪造 UserID
		crc.Reset()
		crc.Write(helper.StringToBytes(event.TinyID))
		r = int64(crc.Sum64() & 0x7fff_ffff_ffff_ffff) // 确保为正数
		if r <= 0xffff_ffff {
			r |= 0x1_0000_0000 // 确保不与正常号码重叠
		}
		event.UserID = r
		if event.Sender != nil {
			event.Sender.ID = r
		}
		msgid = message.NewMessageIDFromString(event.MessageID.(string))
	}

	switch event.PostType { // process DetailType
	case "message", "message_sent":
		event.DetailType = event.MessageType
	case "notice":
		event.DetailType = event.NoticeType
		preprocessNoticeEvent(&event)
	case "request":
		event.DetailType = event.RequestType
	}
	if event.PostType == "message" {
		preprocessMessageEvent(&event)
	}
	ctx := &Ctx{
		Event:  &event,
		State:  State{},
		caller: &messageLogger{msgid: msgid, caller: caller},
	}
	matcherLock.Lock()
	if hasMatcherListChanged {
		matcherListForRanging = make([]*Matcher, len(matcherList))
		copy(matcherListForRanging, matcherList)
		hasMatcherListChanged = false
	}
	matcherLock.Unlock()
	go match(ctx, matcherListForRanging, maxwait)
}

func gorule(ctx *Ctx, rule Rule) <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		defer func() {
			close(ch)
			if pa := recover(); pa != nil {
				log.Errorf("[bot] execute rule err: %v\n%v", pa, helper.BytesToString(debug.Stack()))
			}
		}()
		ch <- rule(ctx)
	}()
	return ch
}

func gohandler(ctx *Ctx, handler Handler) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		defer func() {
			close(ch)
			if pa := recover(); pa != nil {
				log.Errorf("[bot] execute handler err: %v\n%v", pa, helper.BytesToString(debug.Stack()))
			}
		}()
		handler(ctx)
		ch <- struct{}{}
	}()
	return ch
}

func processMatchers(ctx *Ctx, matchers []*Matcher, t *time.Timer) {
	for _, matcher := range matchers {
		if !matcher.Type(ctx) { // 不匹配直接跳过
			continue
		}
		for k := range ctx.State { // Clear State
			delete(ctx.State, k)
		}
		// copy matcher
		m := matcher.copy()
		ctx.ma = m

		// pre handler
		if eng := m.Engine; eng != nil {
			if processEnginePreHandler(ctx, eng, t) { // true 退出循环
				return
			}
		}
		// rules
		if processRules(ctx, m.Rules, t) {
			return
		}
		// mid handler
		if eng := m.Engine; eng != nil {
			if processEngineMidHandler(ctx, eng, t) { // true 退出循环
				return
			}
		}
		// handler
		if processMatcherHandler(ctx, t) { // true 退出循环
			return
		}
		if m.Temp { // 临时 Matcher 删除
			m.Delete()
		}
		// post handler
		if eng := m.Engine; eng != nil {
			if processEnginePostHandler(ctx, eng, t) { // true 退出循环
				return
			}
		}
		if m.Block { // 阻断后续
			return
		}
	}
}

func processRule(ctx *Ctx, rule Rule, t *time.Timer, logStr string) bool {
	c := gorule(ctx, rule)
	for {
		select {
		case ok := <-c:
			if !ok {
				return ctx.ma.Break
			}
		case <-t.C:
			if ctx.ma.NoTimeout {
				t.Reset(BotConfig.MaxProcessTime)
				continue
			}
			log.Warnf("[bot] %v 处理达到最大时延, 退出", logStr)
			return true
		}
		break
	}
	return false
}

func processHandler(ctx *Ctx, handler Handler, t *time.Timer, logStr string) bool {
	c := gohandler(ctx, handler)
	for {
		select {
		case <-c:
		case <-t.C:
			if ctx.ma.NoTimeout {
				t.Reset(BotConfig.MaxProcessTime)
				continue
			}
			log.Warnf("[bot] %v 处理达到最大时延, 退出", logStr)
			return true
		}
		break
	}
	return false
}

func processEnginePreHandler(ctx *Ctx, engine *Engine, t *time.Timer) bool { // 返回是否退出上层循环
	for _, handler := range engine.preHandler {
		if processRule(ctx, handler, t, "preHandler") {
			return true
		}
	}
	return false
}

func processRules(ctx *Ctx, rules []Rule, t *time.Timer) bool {
	for _, rule := range rules {
		if processRule(ctx, rule, t, "rule") {
			return true
		}
	}
	return false
}

func processEngineMidHandler(ctx *Ctx, engine *Engine, t *time.Timer) bool {
	for _, handler := range engine.midHandler {
		if processRule(ctx, handler, t, "midHandler") {
			return true
		}
	}
	return false
}

func processMatcherHandler(ctx *Ctx, t *time.Timer) bool {
	if h := ctx.ma.Handler; h != nil {
		if processHandler(ctx, h, t, "handler") {
			return true
		}
	}
	return false
}

func processEnginePostHandler(ctx *Ctx, engine *Engine, t *time.Timer) bool {
	for _, handler := range engine.postHandler {
		if processHandler(ctx, handler, t, "postHandler") {
			return true
		}
	}
	return false
}

// match 匹配规则，处理事件
func match(ctx *Ctx, matchers []*Matcher, maxwait time.Duration) {
	if BotConfig.MarkMessage && ctx.Event.MessageID != nil {
		ctx.MarkThisMessageAsRead()
	}
	t := time.NewTimer(maxwait)
	defer t.Stop()
	processMatchers(ctx, matchers, t)
}

// preprocessMessageEvent 返回信息事件
func preprocessMessageEvent(e *Event) {
	e.Message = message.ParseMessage(e.NativeMessage)

	processAt := func() { // 处理是否at机器人
		e.IsToMe = false
		for i, m := range e.Message {
			if m.Type == "at" {
				qq, _ := strconv.ParseInt(m.Data["qq"], 10, 64)
				if qq == e.SelfID {
					e.IsToMe = true
					e.Message = append(e.Message[:i], e.Message[i+1:]...)
					return
				}
			}
		}
		if e.Message == nil || len(e.Message) == 0 || e.Message[0].Type != "text" {
			return
		}
		first := e.Message[0]
		first.Data["text"] = strings.TrimLeft(first.Data["text"], " ") // Trim!
		text := first.Data["text"]
		for _, nickname := range BotConfig.NickName {
			if strings.HasPrefix(text, nickname) {
				e.IsToMe = true
				first.Data["text"] = text[len(nickname):]
				return
			}
		}
	}

	switch {
	case e.DetailType == "group":
		log.Infof("[bot] 收到群(%v)消息 %v : %v", e.GroupID, e.Sender.String(), e.RawMessage)
		processAt()
	case e.DetailType == "guild" && e.SubType == "channel":
		log.Infof("[bot] 收到频道(%v)(%v-%v)消息 %v : %v", e.GroupID, e.GuildID, e.ChannelID, e.Sender.String(), e.Message)
		processAt()
	default:
		e.IsToMe = true // 私聊也判断为at
		log.Infof("[bot] 收到私聊消息 %v : %v", e.Sender.String(), e.RawMessage)
	}
	if len(e.Message) > 0 && e.Message[0].Type == "text" { // Trim Again!
		e.Message[0].Data["text"] = strings.TrimLeft(e.Message[0].Data["text"], " ")
	}
}

// preprocessNoticeEvent 更新事件
func preprocessNoticeEvent(e *Event) {
	if e.SubType == "poke" || e.SubType == "lucky_king" {
		e.IsToMe = e.TargetID == e.SelfID
	} else {
		e.IsToMe = e.UserID == e.SelfID
	}
}

// GetBot 获取指定的bot (Ctx)实例
func GetBot(id int64) *Ctx {
	caller, ok := APICallers.Load(id)
	if !ok {
		return nil
	}
	return &Ctx{caller: caller}
}

// RangeBot 遍历所有bot (Ctx)实例
//
// 单次操作返回 true 则继续遍历，否则退出
func RangeBot(iter func(id int64, ctx *Ctx) bool) {
	APICallers.Range(func(key int64, value APICaller) bool {
		return iter(key, &Ctx{caller: value})
	})
}

// GetFirstSuperUser 在 qqs 中获得 SuperUsers 列表的首个 qq
//
// 找不到返回 -1
func (c *Config) GetFirstSuperUser(qqs ...int64) int64 {
	m := make(map[int64]struct{}, len(qqs)*4)
	for _, qq := range qqs {
		m[qq] = struct{}{}
	}
	for _, qq := range c.SuperUsers {
		if _, ok := m[qq]; ok {
			return qq
		}
	}
	return -1
}
