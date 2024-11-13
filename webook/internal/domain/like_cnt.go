package domain

type LikeCnt struct {
	Cnt   int64  `json:"cnt,omitempty"`
	Biz   string `json:"biz,omitempty"`
	BizId int64  `json:"bizId,omitempty"`
}
