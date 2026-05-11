package errno

type ErrNo struct {
	Code int32
	Msg  string
}

func (e ErrNo) Error() string {
	return e.Msg
}

func (e ErrNo) WithMsg(msg string) ErrNo {
	return ErrNo{Code: e.Code, Msg: msg}
}

func New(code int32, msg string) ErrNo {
	return ErrNo{Code: code, Msg: msg}
}

var (
	OK              = New(200, "成功")
	ErrParam        = New(4001, "参数错误")
	ErrNotFound     = New(4002, "记录不存在")
	ErrActivityTime = New(4003, "活动未开始或已结束")
	ErrStockOut     = New(4005, "库存不足")
	ErrRepeat       = New(4006, "请勿重复抢购")
	ErrBusy         = New(4007, "系统繁忙")
	ErrRateLimit    = New(4008, "请求过于频繁")
	ErrUnauthorized = New(4009, "未登录或Token无效")
	ErrInternal     = New(5000, "系统错误")
)
