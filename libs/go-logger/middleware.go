package logger

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const CtxLoggerKey = "ctx_logger"

func FromContext(c echo.Context) *slog.Logger {
	l, ok := c.Get(CtxLoggerKey).(*slog.Logger)
	if !ok {
		return L.Logger
	}
	return l
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Get Request ID
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			// 2. Create sub-logger
			subLogger := L.With("request_id", reqID)

			// 3. Inject into context
			c.Set(CtxLoggerKey, subLogger)

			// 4. Wrap Echo's RequestLogger with a SKIPPER
			return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
				LogStatus:    true,
				LogLatency:   true,
				LogRemoteIP:  true,
				LogMethod:    true,
				LogURI:       true,
				LogUserAgent: true,
				LogError:     true,

				Skipper: func(c echo.Context) bool {
					p := c.Request().URL.Path
					return p == "/health" || p == "/favicon.ico" || p == "/" || p == ""
				},

				LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

					attrs := []any{
						"remote_ip", v.RemoteIP,
						"host", v.Host,
						"method", v.Method,
						"uri", v.URI,
						"status", v.Status,
						"latency", v.Latency,
						"user_agent", v.UserAgent,
					}

					if v.Error != nil {
						subLogger.Error("request summary", append(attrs, "err", v.Error)...)
					} else {
						subLogger.Info("request summary", attrs...)
					}

					return nil
				},
			})(next)(c)
		}
	}
}
