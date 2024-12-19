package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/ljinf/user_auth/common/logger"
	"github.com/ljinf/user_auth/common/util"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// infrastructure 中存放项目运行需要的基础中间价

func StartTrace() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceId := ctx.Request.Header.Get("traceid")
		pSpanId := ctx.Request.Header.Get("spanid")
		spanId := util.GenerateSpanID(ctx.Request.RemoteAddr)

		if traceId == "" {
			// 如果traceId 为空，证明是链路的发端，把它设置成此次的spanId，发端的spanId是root spanId
			traceId = spanId // trace 标识整个请求的链路, span则标识链路中的不同服务
		}

		ctx.Set("traceid", traceId)
		ctx.Set("spanid", spanId)
		ctx.Set("pspanid", pSpanId)
		ctx.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// 包装一下 gin.ResponseWriter，通过这种方式拦截写响应
// 让gin写响应的时候先写到 bodyLogWriter 再写gin.ResponseWriter ，
// 这样利用中间件里输出访问日志时就能拿到响应了
// https://stackoverflow.com/questions/38501325/how-to-log-response-body-in-gin
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogAccess() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//保存body
		reqBody, _ := io.ReadAll(ctx.Request.Body)
		ctx.Request.Body = io.NopCloser(bytes.NewReader(reqBody))
		start := time.Now()
		blw := &bodyLogWriter{
			body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer,
		}
		ctx.Writer = blw
		accessLog(ctx, "access_start", time.Since(start), reqBody, nil)
		defer func() {
			accessLog(ctx, "access_end", time.Since(start), reqBody, blw.body.String())
		}()
		ctx.Next()
		return
	}
}

func accessLog(ctx *gin.Context, accessType string, duration time.Duration,
	body []byte, dataOut interface{}) {

	req := ctx.Request
	bodyStr := string(body)
	query := req.URL.RawQuery
	path := req.URL.Path
	// TODO: 实现Token认证后再把访问日志里也加上token记录
	// token := c.Request.Header.Get("token")
	logger.New().Info(ctx, "AccessLog",
		"type", accessType,
		"ip", ctx.ClientIP(),
		//"token",token,
		"method", req.Method,
		"path", path,
		"query", query,
		"body", bodyStr,
		"output", dataOut,
		"time(ms)", int64(duration/time.Millisecond),
	)
}

// GinPanicRecovery 自定义gin recover输出
func GinPanicRecovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				if brokenPipe {
					logger.New().Error(ctx, "http request broken pipe", "path", ctx.Request.URL.Path, "error", err, "request", string(httpRequest))
					// If the connection is dead, we can't write a status to it.
					ctx.Error(err.(error)) // nolint: errcheck
					ctx.Abort()
					return
				}

				logger.New().Error(ctx, "http_request_panic", "path", ctx.Request.URL.Path, "error", err, "request", string(httpRequest), "stack", string(debug.Stack()))

				ctx.AbortWithError(http.StatusInternalServerError, err.(error))
			}
		}()
		ctx.Next()
	}
}
