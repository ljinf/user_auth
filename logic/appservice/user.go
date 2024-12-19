package appservice

import (
	"context"
	"github.com/ljinf/user_auth/api/reply"
	"github.com/ljinf/user_auth/common/logger"
	"github.com/ljinf/user_auth/common/util"
	"github.com/ljinf/user_auth/logic/domainservice"
)

type UserAppSvc struct {
	ctx           context.Context
	userDomainSvc *domainservice.UserDomainSvc
}

func NewUserAppSvc(ctx context.Context) *UserAppSvc {
	return &UserAppSvc{
		ctx:           ctx,
		userDomainSvc: domainservice.NewUserDomainSvc(ctx),
	}
}

func (us *UserAppSvc) GenToken() (*reply.TokenReply, error) {
	token, err := us.userDomainSvc.GenAuthToken(12345678, "h5", "")
	if err != nil {
		return nil, err
	}
	logger.New().Info(us.ctx, "generate token success", "tokenData", token)
	tokenReply := new(reply.TokenReply)
	util.CopyProperties(tokenReply, token)
	return tokenReply, err
}

func (us *UserAppSvc) TokenRefresh(refreshToken string) (*reply.TokenReply, error) {
	token, err := us.userDomainSvc.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	logger.New().Info(us.ctx, "refresh token success", "tokenData", token)
	tokenReply := new(reply.TokenReply)
	util.CopyProperties(tokenReply, token)
	return tokenReply, err
}
