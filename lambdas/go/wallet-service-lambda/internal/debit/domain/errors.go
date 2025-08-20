package domain

type Error struct {
	Message  string
	Code     string
	Cause    error
	Metadata map[string]any
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

func NewInsufficientFundsError(id string, available, requested float64) error {
	return &Error{
		Message: "insufficient funds error",
		Code:    "4001",
		Metadata: map[string]any{
			"id":               id,
			"availableBalance": available,
			"requestedAmount":  requested},
	}
}

func NewMaxRetriesError(id string, e error) error {
	return &Error{
		Message:  "max retries exceeded error",
		Code:     "4002",
		Cause:    e,
		Metadata: map[string]any{"id": id},
	}
}

func NewGetFundsError(id string, e error) error {
	return &Error{
		Message:  "get funds error",
		Code:     "5001",
		Cause:    e,
		Metadata: map[string]any{"id": id},
	}
}

func NewDebitFundsError(id string, e error) error {
	return &Error{
		Message:  "debit funds error",
		Code:     "5002",
		Cause:    e,
		Metadata: map[string]any{"id": id},
	}
}

func NewPublishMessageError(id string, e error) error {
	return &Error{
		Message:  "debit funds error",
		Code:     "5003",
		Cause:    e,
		Metadata: map[string]any{"id": id},
	}
}
