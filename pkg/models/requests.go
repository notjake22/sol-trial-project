package models

type GenericResponse[T any] struct {
	Object  T      `json:"object"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}
