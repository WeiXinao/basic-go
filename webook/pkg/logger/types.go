package logger

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func LoggerExample() {
	var l Logger
	phone := "152xxxx1234"
	l.Info("用户未注册，手机号码是 %s", phone)
}

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
	// args 会加入进去 LoggerV1 的任何打印出来的日志里面
	With(args ...Field) LoggerV1
}

type Field struct {
	Key   string
	Value any
}

func LoggerV1Example() {
	var l LoggerV1
	phone := "152xxxx1234"
	l.Info("用户未注册", Field{
		Key:   "phone",
		Value: phone,
	})
}

type LoggerV2 interface {
	// args 必须是偶数，并且按照 key-value, key-value 来组织
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func LoggerV2Example() {
	var l LoggerV2
	phone := "152xxxx1234"
	l.Info("用户未注册", "phone", phone)
}
