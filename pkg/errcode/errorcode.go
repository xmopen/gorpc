package errcode

type Error *ErrorMessage

// ErrorMessage go_rpc内定义的Error.
type ErrorMessage struct {
	Code    int64
	Message string
}

// NewError return a ErrorMessage instance
func NewError(code int64, msg string) *ErrorMessage {
	return &ErrorMessage{
		Code:    code,
		Message: msg,
	}
}
