package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ljinf/user_auth/common/enum"
	"github.com/ljinf/user_auth/common/errcode"
	"github.com/ljinf/user_auth/common/logger"
	"github.com/ljinf/user_auth/logic/do"
	"github.com/redis/go-redis/v9"
	"time"
)

// SetUserToken 设置用户的AccessToken 和 RefreshToken 缓存
func SetUserToken(ctx context.Context, session *do.SessionInfo) error {
	log := logger.New()
	err := setAccessToken(ctx, session)
	if err != nil {
		log.Error(ctx, "redis error", "err", err)
		return err
	}
	err = setRefreshToken(ctx, session)
	if err != nil {
		log.Error(ctx, "redis error", "err", err)
		return err
	}
	return err
}

func SetUserSession(ctx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, session.UserId)
	sessionDataBytes, _ := json.Marshal(session)
	err := Redis().HSet(ctx, redisKey, session.Platform, sessionDataBytes).Err()
	if err != nil {
		logger.New().Error(ctx, "redis error", "err", err)
		return err
	}
	return err
}

// DelOldSessionTokens 删除用户旧Session的Token
func DelOldSessionTokens(ctx context.Context, session *do.SessionInfo) error {
	//log := logger.New(ctx)
	oldSession, err := GetUserPlatformSession(ctx, session.UserId, session.Platform)
	if err != nil {
		return err
	}
	if oldSession == nil {
		// 没有旧Session
		return nil
	}
	err = DelAccessToken(ctx, oldSession.AccessToken)
	if err != nil {
		return errcode.Wrap("redis error", err)
	}
	err = DelayDelRefreshToken(ctx, oldSession.RefreshToken)
	if err != nil {
		return errcode.Wrap("redis error", err)
	}
	return nil
}

// GetUserPlatformSession 获取用户在指定平台中的Session信息
func GetUserPlatformSession(ctx context.Context, userId int64, platform string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_USER_SESSION, userId)
	result, err := Redis().HGet(ctx, redisKey, platform).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	// key 不存在
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	session := new(do.SessionInfo)
	err = json.Unmarshal([]byte(result), &session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func setAccessToken(ctx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, session.AccessToken)
	sessionDataBytes, _ := json.Marshal(session)
	res, err := Redis().Set(ctx, redisKey, sessionDataBytes, enum.AccessTokenDuration).Result()
	logger.New().Debug(ctx, "redis debug", "res", res, "err", err)
	return err
}

func setRefreshToken(ctx context.Context, session *do.SessionInfo) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, session.RefreshToken)
	sessionDataBytes, _ := json.Marshal(session)
	return Redis().Set(ctx, redisKey, sessionDataBytes, enum.RefreshTokenDuration).Err()
}

func DelAccessToken(ctx context.Context, accessToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, accessToken)
	return Redis().Del(ctx, redisKey).Err()
}

// DelayDelRefreshToken 刷新Token时让旧的RefreshToken 保留一段时间自己过期
func DelayDelRefreshToken(ctx context.Context, refreshToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	return Redis().Expire(ctx, redisKey, enum.OldRefreshTokenHoldingDuration).Err()
}

// DelRefreshToken 直接删除RefreshToken缓存  修改密码、退出登录时使用
func DelRefreshToken(ctx context.Context, refreshToken string) error {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	return Redis().Del(ctx, redisKey).Err()
}

// TODO Delete user's session of specific platform

// TODO Delete user's session of all platform

func LockTokenRefresh(ctx context.Context, refreshToken string) (bool, error) {
	redisLockKey := fmt.Sprintf(enum.REDISKEY_TOKEN_REFRESH_LOCK, refreshToken)
	return Redis().SetNX(ctx, redisLockKey, "locked", 10*time.Second).Result()
}

func UnlockTokenRefresh(ctx context.Context, refreshToken string) error {
	redisLockKey := fmt.Sprintf(enum.REDISKEY_TOKEN_REFRESH_LOCK, refreshToken)
	return Redis().Del(ctx, redisLockKey).Err()
}

func GetRefreshToken(ctx context.Context, refreshToken string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_REFRESH_TOKEN, refreshToken)
	result, err := Redis().Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	session := new(do.SessionInfo)
	if errors.Is(err, redis.Nil) {
		return session, nil
	}
	json.Unmarshal([]byte(result), &session)

	return session, nil
}

func GetAccessToken(ctx context.Context, accessToken string) (*do.SessionInfo, error) {
	redisKey := fmt.Sprintf(enum.REDIS_KEY_ACCESS_TOKEN, accessToken)
	result, err := Redis().Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	session := new(do.SessionInfo)
	if errors.Is(err, redis.Nil) {
		return session, nil
	}
	json.Unmarshal([]byte(result), &session)

	return session, nil
}
