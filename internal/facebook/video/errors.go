package video

// FacebookError represents a Facebook API error
type FacebookError struct {
	Message      string                 `json:"message"`
	Type         string                 `json:"type"`
	Code         int                    `json:"code"`
	ErrorSubcode int                    `json:"error_subcode"`
	IsTransient  bool                   `json:"is_transient"`
	ErrorData    map[string]interface{} `json:"error_data"`
}

// Error implements the error interface
func (e *FacebookError) Error() string {
	return e.Message
}