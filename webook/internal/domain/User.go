package domain

import "time"

// User 领域对象，是 DDD 中的聚合根，也可以叫 entity
// BO(business object)
type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	Nickname string
	Birthday time.Time
	AboutMe  string

	Ctime time.Time
}
