// Package domain defines the core business entities and value objects for the RTSP surveillance system.
package domain

import "fmt"

// Error types for domain-level operations.

// NotFoundError indicates that a requested resource was not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ValidationError indicates that input data failed validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// ConflictError indicates that the operation conflicts with existing data.
type ConflictError struct {
	Resource string
	Reason   string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Reason)
}

// UnauthorizedError indicates that the request is not authorized.
type UnauthorizedError struct {
	Reason string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Reason)
}

// InternalError indicates an unexpected internal server error.
type InternalError struct {
	Operation string
	Err       error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error during %s: %v", e.Operation, e.Err)
}
