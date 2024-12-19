package app

import (
	"github.com/gin-gonic/gin"
	"github.com/ljinf/user_auth/common/errcode"
	"github.com/ljinf/user_auth/common/logger"
)

type response struct {
	ctx        *gin.Context
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	RequestId  string      `json:"request_id"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *pagination `json:"pagination,omitempty"`
}

func NewResponse(c *gin.Context) *response {
	return &response{ctx: c}
}

// SetPagination 设置Response的分页信息
func (r *response) SetPagination(pagination *pagination) *response {
	r.Pagination = pagination
	return r
}

func (r *response) Success(data interface{}) {
	r.Code = errcode.Success.Code()
	r.Msg = errcode.Success.Msg()
	requestId := ""
	if _, exists := r.ctx.Get("traceid"); exists {
		val, _ := r.ctx.Get("traceid")
		requestId = val.(string)
	}
	r.RequestId = requestId
	r.Data = data

	r.ctx.JSON(errcode.Success.HttpStatusCode(), r)
}

func (r *response) SuccessOk() {
	r.Success("")
}

func (r *response) Error(err *errcode.AppError) {
	r.Code = err.Code()
	r.Msg = err.Msg()
	if _, exists := r.ctx.Get("traceid"); exists {
		val, _ := r.ctx.Get("traceid")
		r.RequestId = val.(string)
	}
	// 兜底记一条响应错误, 项目自定义的AppError中有错误链条, 方便出错后排查问题
	logger.New().Error(r.ctx, "api_response_error", "err", err)
	r.ctx.JSON(err.HttpStatusCode(), r)
}
