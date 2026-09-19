package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type UserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogic) GetCurrentUser() (*types.UserResp, error) {
	name := "Su Huihui"
	jwtIssuer := "https://subuhui.internal"

	var user db.User
	if err := l.svcCtx.DB.First(&user).Error; err != nil {
		return &types.UserResp{
			ID:                    1,
			Email:                 "admin@subuhui.internal",
			DisplayName:           &name,
			HasActiveSubscription: true,
			JWTIssuer:             jwtIssuer,
		}, nil
	}

	userName := user.Name
	if userName == "" {
		userName = name
	}

	return &types.UserResp{
		ID:                    user.ID,
		Email:                 user.Email,
		DisplayName:           &userName,
		HasActiveSubscription: user.HasActiveSubscription,
		JWTIssuer:             jwtIssuer,
	}, nil
}
