package http_utils

import (
	"encoding/json"
	"io"
	"net/http"
)

func ReadSerializable(r *http.Request, item any) error {
	buff, _ := io.ReadAll(r.Body)

	err := json.Unmarshal(buff, item)

	return err
}

func ReadSerializableFromResponse(r *http.Response, item any) error {
	buff, _ := io.ReadAll(r.Body)

	err := json.Unmarshal(buff, item)

	return err
}

func WriteSerializable(item any, w http.ResponseWriter) {
	bytes, _ := json.Marshal(item)
	w.Write(bytes)
}

func HealthCkeck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
