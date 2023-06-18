package errcode

var (
	RegisterNoServiceForType = &ErrorMessage{
		Code:    1001,
		Message: "go_rpc.RegisterErr: no service name for type",
	}

	RegisterServerNameNoExported = &ErrorMessage{
		Code:    1002,
		Message: "go_rpc.RegisterErr: server name no exported",
	}
)
