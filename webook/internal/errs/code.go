package errs

// User 相关
const (
	// UserInvalidInput 统一的用户模块输入错误
	UserInvalidInput = 401001 + iota
	// UserInvalidOrPassword 用户名或密码不对
	UserInvalidOrPassword
	//	UserDuplicateEmail 用户邮箱冲突
	UserDuplicateEmail
	UserInternalServerError = 501001
)

const (
	//	ArticleInvalidInput 文章模块的统一的错误码
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)

type BizErr struct {
	Code uint
	Msg  string
}
