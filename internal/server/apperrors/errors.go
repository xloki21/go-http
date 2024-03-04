package apperrors

import "net/http"

var (
	MethodNotAllowed    = AppError{Code: http.StatusMethodNotAllowed, Message: "Method not allowed"}
	InvalidBodyErr      = AppError{Code: http.StatusBadRequest, Message: "Invalid Request. Unable to decode request body"}
	BadRequestErr       = AppError{Code: http.StatusBadRequest, Message: "Bad Request"}
	EmptyListErr        = AppError{Code: http.StatusBadRequest, Message: "Invalid Request. Empty URL list"}
	TooBigURLListErr    = AppError{Code: http.StatusBadRequest, Message: "Invalid Request. URL list exceeds maximum length"}
	InternalErr         = AppError{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
	Teapot              = AppError{Code: http.StatusTeapot, Message: "Teapot"}
	NilErr              = AppError{Code: http.StatusOK, Message: ""}
	TooManyRequestsErr  = AppError{Code: http.StatusTooManyRequests, Message: "Too Many Requests"}
	TimeoutErr          = AppError{Code: http.StatusRequestTimeout, Message: "Request Timeout"}
	RequestCancelledErr = AppError{Code: http.StatusRequestTimeout, Message: "Request Cancelled"}
	URLNotFoundErr      = AppError{Code: http.StatusBadRequest, Message: "One or more URLs were not found"}
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (a AppError) Error() string {
	return a.Err.Error()
}
