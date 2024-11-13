package service

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
)

type LikeCntService interface {
	LikeCntTop100(ctx context.Context) error
}

type likeCntService struct {
	likeCntRepo repository.LikeCntRepository
}

func (l *likeCntService) LikeCntTop100(ctx context.Context) error {
	return l.likeCntRepo.LikeCntTop100(ctx)
}
