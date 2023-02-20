package errcode

var (
	UsernameHasExisted      = NewError(20001, "用户名已存在")
	UsernameLengthLimit     = NewError(20002, "用户名长度3~12")
	UsernameCharLimit       = NewError(20003, "用户名只能包含字母、数字")
	PasswordLengthLimit     = NewError(20004, "密码长度6~16")
	UserRegisterFailed      = NewError(20005, "用户注册失败")
	UserHasBeenBanned       = NewError(20006, "该账户已被封停")
	NoPermission            = NewError(20007, "无权限执行该请求")
	UserHasBindOTP          = NewError(20008, "当前用户已绑定二次验证")
	UserOTPInvalid          = NewError(20009, "二次验证码验证失败")
	UserNoBindOTP           = NewError(20010, "当前用户未绑定二次验证")
	ErrorOldPassword        = NewError(20011, "当前用户密码验证失败")
	ErrorCaptchaPassword    = NewError(20012, "图形验证码验证失败")
	AccountNoPhoneBind      = NewError(20013, "拒绝操作: 账户未绑定手机号")
	TooManyLoginError       = NewError(20014, "登录失败次数过多，请稍后再试")
	GetPhoneCaptchaError    = NewError(20015, "短信验证码获取失败")
	TooManyPhoneCaptchaSend = NewError(20016, "短信验证码获取次数已达今日上限")
	ExistedUserPhone        = NewError(20017, "该手机号已被绑定")
	ErrorPhoneCaptcha       = NewError(20018, "手机验证码不正确")
	MaxPhoneCaptchaUseTimes = NewError(20019, "手机验证码已达最大使用次数")
	NicknameLengthLimit     = NewError(20020, "昵称长度2~12")
	NoExistUserAddress      = NewError(20021, "用户不存在")
	NoAdminPermission       = NewError(20022, "无管理权限")

	GetPostsFailed          = NewError(30001, "获取动态列表失败")
	CreatePostFailed        = NewError(30002, "动态发布失败")
	GetPostFailed           = NewError(30003, "获取动态详情失败")
	DeletePostFailed        = NewError(30004, "动态删除失败")
	LockPostFailed          = NewError(30005, "动态锁定失败")
	GetPostTagsFailed       = NewError(30006, "获取话题列表失败")
	InvalidDownloadReq      = NewError(30007, "附件下载请求不合法")
	DownloadReqError        = NewError(30008, "附件下载请求失败")
	InsuffientDownloadMoney = NewError(30009, "附件下载失败:账户资金不足")
	DownloadExecFail        = NewError(30010, "附件下载失败:扣费失败")
	StickPostFailed         = NewError(30011, "动态置顶失败")
	VisblePostFailed        = NewError(30012, "更新可见性失败")

	GetCommentsFailed   = NewError(40001, "获取评论列表失败")
	CreateCommentFailed = NewError(40002, "评论发布失败")
	GetCommentFailed    = NewError(40003, "获取评论详情失败")
	DeleteCommentFailed = NewError(40004, "评论删除失败")
	CreateReplyFailed   = NewError(40005, "评论回复失败")
	GetReplyFailed      = NewError(40006, "获取评论详情失败")
	MaxCommentCount     = NewError(40007, "评论数已达最大限制")

	GetMessagesFailed = NewError(50001, "获取消息列表失败")
	ReadMessageFailed = NewError(50002, "标记消息已读失败")
	SendWhisperFailed = NewError(50003, "私信发送失败")
	NoWhisperToSelf   = NewError(50004, "不允许给自己发送私信")
	TooManyWhisperNum = NewError(50005, "今日私信次数已达上限")

	GetCollectionsFailed = NewError(60001, "获取收藏列表失败")
	GetStarsFailed       = NewError(60002, "获取点赞列表失败")

	RechargeReqFail     = NewError(70001, "充值请求失败")
	RechargeNotifyError = NewError(70002, "充值回调失败")
	GetRechargeFailed   = NewError(70003, "充值详情获取失败")

	CreateDaoFailed = NewError(80001, "DAO创建失败")
)
