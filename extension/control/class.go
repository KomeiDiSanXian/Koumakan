package control

// Option is the option of Control
type Option[Ctx any] struct {
	DefaultDisable    bool
	Extra             int16         // 额外数据
	Brief             string        // 简介
	HelpText          string        // 帮助文本
	Banner            string        // 背景图
	PrivateDataFolder string        // 私有数据文件夹
	PublicDataFolder  string        // 公共数据文件夹
	OnEnable          func(ctx Ctx) // 启用时的回调
	OnDisable         func(ctx Ctx) // 禁用时的回调
}

// BlockStatus 全局 ban 某人
type BlockStatus struct {
	UserID int64 `db:"uid"`
}

// ResponseGroup 响应的群
type ResponseGroup struct {
	GroupID int64  `db:"gid"` // GroupID 群号, 个人为负
	Extra   string `db:"ext"` // Extra 该群的扩展数据
}

// GroupConfig holds the group config for the Manager.
type GroupConfig struct {
	GroupID int64 `db:"gid"`     // GroupID 群号
	Disable int64 `db:"disable"` // Disable 默认启用该插件
}

// BanStatus 在某群封禁某人的状态
type BanStatus struct {
	ID      int64 `db:"id"`
	UserID  int64 `db:"uid"`
	GroupID int64 `db:"gid"`
}
