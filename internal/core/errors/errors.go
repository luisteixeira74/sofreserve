package errors

var (
	ErrInvalidEventID = &AppError{
		Message: "Evento inválido",
		Code:    "INVALID_EVENT_ID",
	}

	ErrInvalidName = &AppError{
		Message: "Nome é obrigatório",
		Code:    "INVALID_NAME",
	}

	ErrInvalidEmail = &AppError{
		Message: "Email é obrigatório",
		Code:    "INVALID_EMAIL",
	}

	ErrInvalidQuantity = &AppError{
		Message: "Quantidade inválida",
		Code:    "INVALID_QUANTITY",
	}

	ErrEmailAlreadyUsed = &AppError{
		Message: "Este email já possui reserva para este evento",
		Code:    "DUPLICATE_RESERVATION",
	}

	ErrNotEnoughSeats = &AppError{
		Message: "Não há vagas disponíveis",
		Code:    "NO_SEATS_AVAILABLE",
	}

	ErrEventNotFound = &AppError{
		Message: "Evento não encontrado",
		Code:    "EVENT_NOT_FOUND",
	}

	ErrEventClosed = &AppError{
		Message: "As reservas para este evento estão encerradas",
		Code:    "EVENT_CLOSED",
	}
)

type AppError struct {
	Message string
	Code    string
}

func (e *AppError) Error() string {
	return e.Message
}