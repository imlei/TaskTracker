package api

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"simpletask/internal/store"
)

// ErrorResponse 标准化的错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// 常见错误代码
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeBadInput         = "BAD_INPUT"
	ErrCodeTaskLocked       = "TASK_LOCKED"
	ErrCodeCustomerInactive = "CUSTOMER_INACTIVE"
)

// APIError API 错误接口
type APIError struct {
	HTTPCode int
	Code     string
	Message  string
	Err      error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewAPIError 创建新的 API 错误
func NewAPIError(httpCode int, code, message string, err error) *APIError {
	return &APIError{
		HTTPCode: httpCode,
		Code:     code,
		Message:  message,
		Err:      err,
	}
}

// HandleError 统一错误处理函数
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// 记录错误日志
	log.Printf("Error: %v\nStack: %s", err, debug.Stack())

	var apiErr *APIError
	var response ErrorResponse

	// 根据错误类型确定响应
	switch {
	case err == store.ErrNotFound:
		apiErr = NewAPIError(http.StatusNotFound, ErrCodeNotFound, "Resource not found", err)
	case err == store.ErrTaskLocked:
		apiErr = NewAPIError(http.StatusConflict, ErrCodeTaskLocked, "Task is locked and cannot be modified", err)
	case err == store.ErrTaskPaidLocked:
		apiErr = NewAPIError(http.StatusConflict, ErrCodeTaskLocked, "Paid task cannot be modified", err)
	case err == store.ErrTaskDeleteLocked:
		apiErr = NewAPIError(http.StatusConflict, ErrCodeTaskLocked, "Only pending tasks can be deleted", err)
	case err == store.ErrCustomerInactive:
		apiErr = NewAPIError(http.StatusConflict, ErrCodeCustomerInactive, "Customer is inactive", err)
	case strings.Contains(err.Error(), "invalid email"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	case strings.Contains(err.Error(), "invalid phone"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	case strings.Contains(err.Error(), "cannot be empty"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	case strings.Contains(err.Error(), "too long"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	case strings.Contains(err.Error(), "invalid password"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	case strings.Contains(err.Error(), "invalid amount"):
		apiErr = NewAPIError(http.StatusBadRequest, ErrCodeValidation, err.Error(), err)
	default:
		apiErr = NewAPIError(http.StatusInternalServerError, ErrCodeInternal, "Internal server error", err)
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPCode)

	// 构建响应
	response = ErrorResponse{
		Error:   apiErr.Code,
		Message: apiErr.Message,
	}

	// 开发环境下可以包含详细错误信息
	// if isDevelopment() {
	//     response.OriginalError = apiErr.Err.Error()
	// }

	// 返回 JSON 响应
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// RecoverMiddleware 恢复中间件，捕获 panic
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\nStack: %s", err, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := ErrorResponse{
					Error:   ErrCodeInternal,
					Message: "Internal server error",
				}

				if err := json.NewEncoder(w).Encode(response); err != nil {
					log.Printf("Failed to encode panic response: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := r.Context().Value("request_start")
		if start == nil {
			// 记录请求信息
			log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}

// CORSHeadersMiddleware 添加 CORS 头
func CORSHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取允许的源（从环境变量或配置中读取）
		allowedOrigins := []string{"*"} // 默认允许所有，生产环境应该配置

		origin := r.Header.Get("Origin")
		if origin != "" {
			// 检查是否在允许列表中
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ResponseWriterWrapper 用于包装 http.ResponseWriter 以捕获状态码
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			if err := recover(); err != nil {
				HandleError(w, err.(error))
			}
		}()

		next.ServeHTTP(wrapped, r)
	})
}
