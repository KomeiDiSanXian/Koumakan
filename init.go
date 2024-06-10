package zero

func init() {
	defaultEngine.UsePreHandler(
		func(ctx *Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
	)
}