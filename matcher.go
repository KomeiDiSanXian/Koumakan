package koumakan

import (
	"sort"
	"sync"
)

type (
	// Rule filter the event
	Rule func(ctx *Ctx) bool
	// Handler 事件处理函数
	Handler func(ctx *Ctx)
)

// Matcher 是 ZeroBot 匹配和处理事件的最小单元
type Matcher struct {
	// Temp 是否为临时Matcher，临时 Matcher 匹配一次后就会删除当前 Matcher
	Temp bool
	// Block 是否阻断后续 Matcher，为 true 时当前Matcher匹配成功后，后续Matcher不参与匹配
	Block bool
	// Break 是否退出后续匹配流程, 只有 rule 返回 false 且此值为真才会退出, 且不对 mid handler 以下的 rule 生效
	Break bool
	// NoTimeout 处理是否不设超时
	NoTimeout bool
	// Priority 优先级，越小优先级越高
	Priority int
	// Event 当前匹配到的事件
	// Event *Event // Event 更多的是从 Ctx 中获取，removed
	// Type 匹配的事件类型
	Type Rule
	// Rules 匹配规则
	Rules []Rule
	// Handler 处理事件的函数
	Handler Handler
	// Engine 注册 Matcher 的 Engine，Engine可为一系列 Matcher 添加通用 Rule 和 其他钩子
	Engine Engine
}

var (
	// 所有主匹配器列表
	matcherList = make([]IMatcher, 0)
	// Matcher 修改读写锁
	matcherLock = sync.RWMutex{}
	// 用于迭代的所有主匹配器列表
	matcherListForRanging []IMatcher
	// 是否 matcherList 已经改变
	// 如果改变，下次迭代需要更新
	// matcherListForRanging
	hasMatcherListChanged bool
)

// State store the context of a matcher.
type State map[string]any

// GetPriority 获取当前 Matcher 优先级
func (m *Matcher) GetPriority() int {
	return m.Priority
}

// IsTemp 是否为临时Matcher
func (m *Matcher) IsTemp() bool {
	return m.Temp
}

// GetType 获取当前 Matcher 匹配的事件类型
func (m *Matcher) GetType() Rule {
	return m.Type
}

// IsBlock 是否阻断后续 Matcher
func (m *Matcher) IsBlock() bool {
	return m.Block
}

// IsBreak 是否退出后续匹配流程
func (m *Matcher) IsBreak() bool {
	return m.Break
}

// IsNoTimeout 是否不设超时
func (m *Matcher) IsNoTimeout() bool {
	return m.NoTimeout
}

// GetRules 获取当前 Matcher 的匹配规则
func (m *Matcher) GetRules() []Rule {
	return m.Rules
}

// GetHandler 获取当前 Matcher 的处理函数
func (m *Matcher) GetHandler() Handler {
	return m.Handler
}

// GetEngine 获取当前 Matcher 的 Engine
func (m *Matcher) GetEngine() Engine {
	return m.Engine
}

func sortMatcher() {
	sort.Slice(matcherList, func(i, j int) bool { // 按优先级排序
		return matcherList[i].GetPriority() < matcherList[j].GetPriority()
	})
	hasMatcherListChanged = true
}

// SetBlock 设置是否阻断后面的 Matcher 触发
func (m *Matcher) SetBlock(block bool) IMatcher {
	m.Block = block
	return m
}

// SetPriority 设置当前 Matcher 优先级
func (m *Matcher) SetPriority(priority int) IMatcher {
	matcherLock.Lock()
	defer matcherLock.Unlock()
	m.Priority = priority
	sortMatcher()
	return m
}

// BindEngine bind the matcher to a engine
func (m *Matcher) BindEngine(e Engine) IMatcher {
	m.Engine = e
	return m
}

// StoreMatcher store a matcher to matcher list.
func StoreMatcher(m *Matcher) IMatcher {
	matcherLock.Lock()
	defer matcherLock.Unlock()
	// todo(wdvxdr): move to engine.
	if m.Engine != nil {
		m.Block = m.Block || m.Engine.getBlock()
	}
	matcherList = append(matcherList, m)
	sortMatcher()
	return m
}

// StoreTempMatcher store a matcher only triggered once.
func StoreTempMatcher(m *Matcher) IMatcher {
	m.Temp = true
	StoreMatcher(m)
	return m
}

// Delete remove the matcher from list
func (m *Matcher) Delete() {
	matcherLock.Lock()
	defer matcherLock.Unlock()
	for i, matcher := range matcherList {
		if m == matcher {
			matcherList = append(matcherList[:i], matcherList[i+1:]...)
			hasMatcherListChanged = true
		}
	}
}

func (m *Matcher) copy() IMatcher {
	return &Matcher{
		Type:     m.Type,
		Rules:    m.Rules,
		Block:    m.Block,
		Priority: m.Priority,
		Handler:  m.Handler,
		Temp:     m.Temp,
		Engine:   m.Engine,
	}
}

// Handle 直接处理事件
func (m *Matcher) Handle(handler Handler) IMatcher {
	m.Handler = handler
	return m
}
