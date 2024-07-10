package zero

// APICallerReturnHook is a caller middleware
type APICallerReturnHook struct {
	caller   APICaller
	callback func(rsp APIResponse, err error)
}

// NewAPICallerReturnHook hook ctx's caller
func NewAPICallerReturnHook(ctx *Ctx, callback func(rsp APIResponse, err error)) (v *APICallerReturnHook) {
	return &APICallerReturnHook{
		caller:   ctx.caller,
		callback: callback,
	}
}

// CallApi call original caller and pass rsp to callback
func (v *APICallerReturnHook) CallApi(request APIRequest) (rsp APIResponse, err error) {
	rsp, err = v.caller.CallApi(request)
	go v.callback(rsp, err)
	return
}
