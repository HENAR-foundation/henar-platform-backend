package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"henar-backend/sentry"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MalformedRequest struct {
	Status int
	Msg    string
}

func (mr *MalformedRequest) Error() string {
	return mr.Msg
}

func DecodeJSONBody(c *fiber.Ctx, dst interface{}) error {
	if c.Get("Content-Type") != "application/json" {
		msg := "Content-Type header is not application/json"
		return &MalformedRequest{Status: http.StatusUnsupportedMediaType, Msg: msg}
	}

	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		sentry.SentryHandler(err)
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf(
				"Request body contains badly-formed JSON (at position %d)",
				syntaxError.Offset,
			)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return &MalformedRequest{
				Status: http.StatusBadRequest,
				Msg:    "Request body contains badly-formed JSON",
			}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf(
				"Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field,
				unmarshalTypeError.Offset,
			)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &MalformedRequest{Status: http.StatusRequestEntityTooLarge, Msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		sentry.SentryHandler(err)
		msg := "Request body must only contain a single JSON object"
		return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}
	}

	return nil
}

func ConvertStringArrayToObjectIDArray(stringArray []string) ([]primitive.ObjectID, error) {
	objectIDArray := make([]primitive.ObjectID, len(stringArray))

	for i, str := range stringArray {
		objectID, err := primitive.ObjectIDFromHex(str)
		if err != nil {
			sentry.SentryHandler(err)
			return nil, err
		}
		objectIDArray[i] = objectID
	}

	return objectIDArray, nil
}
