package http

type ApiResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
	Path    string `json:"path,omitempty"`
}

func Response(message string, data any, error any, path string) *ApiResponse {
	return &ApiResponse{
		Message: message,
		Data:    data,
		Error:   error,
		Path:    path,
	}
}
