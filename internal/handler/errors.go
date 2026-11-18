package handler

import (
	"encoding/json"
	"net/http"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// ErrorConverter converts domain errors to HTTP status codes and messages.
type ErrorConverter struct{}

// NewErrorConverter creates a new error converter.
func NewErrorConverter() *ErrorConverter {
	return &ErrorConverter{}
}

// ToHTTPStatus converts a domain error to an HTTP status code.
func (ec *ErrorConverter) ToHTTPStatus(err error) int {
	switch err.(type) {
	case *domain.NotFoundError:
		return http.StatusNotFound
	case *domain.ValidationError:
		return http.StatusBadRequest
	case *domain.ConflictError:
		return http.StatusConflict
	case *domain.UnauthorizedError:
		return http.StatusUnauthorized
	case *domain.InternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ToHTTPMessage converts a domain error to an HTTP error message.
func (ec *ErrorConverter) ToHTTPMessage(err error) string {
	if err == nil {
		return ""
	}

	switch e := err.(type) {
	case *domain.NotFoundError:
		return e.Error()
	case *domain.ValidationError:
		return e.Error()
	case *domain.ConflictError:
		return e.Error()
	case *domain.UnauthorizedError:
		return e.Error()
	case *domain.InternalError:
		return e.Error()
	default:
		return "internal server error"
	}
}

// HandleDomainError writes a domain error as an HTTP response.
func (ec *ErrorConverter) HandleDomainError(w http.ResponseWriter, err error) {
	status := ec.ToHTTPStatus(err)
	message := ec.ToHTTPMessage(err)

	errorHandler := &ErrorHandler{}
	errorHandler.WriteError(w, status, message)
}

// ErrorHandler writes HTTP error responses.
type ErrorHandler struct{}

// WriteError writes an error response in JSON format.
func (eh *ErrorHandler) WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)

	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}
