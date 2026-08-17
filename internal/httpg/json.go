package httpg

// Общая логика работы с JSON
import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

const maxRequestBodySize = 1 << 20 // 1 MiB

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func DecodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	target any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		// На этом этапе статус и часть ответа уже могли быть отправлены,
		// поэтому клиенту другой HTTP-ответ вернуть невозможно.
		log.Printf("encode HTTP response: %v", err)
	}
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	WriteJSON(w, status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
}

func WriteInvalidRequest(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
}

func WriteValidationError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, "validation_error", err.Error())
}

func WriteForbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, "forbidden", message)
}

func WriteInternalError(
	w http.ResponseWriter,
	operation string,
	err error,
) {
	log.Printf("%s: %v", operation, err)

	WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
