package http_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lapayka/rsoi-2/Common/Logger"
)

func ReadSerializable(r *http.Request, item any) error {
	buff, err := io.ReadAll(r.Body)

	fmt.Println(string(buff))

	if err != nil {
		Logger.GetLogger().Print(err)
		return err
	}

	err = json.Unmarshal(buff, item)

	return err
}

func ReadSerializableFromResponse(r *http.Response, item any) error {
	buff, _ := io.ReadAll(r.Body)
	fmt.Println(string(buff))

	err := json.Unmarshal(buff, item)

	return err
}

func WriteSerializable(item any, w http.ResponseWriter) {
	bytes, _ := json.Marshal(item)
	w.Write(bytes)
}

func HealthCkeck(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Checked Health")
	w.WriteHeader(http.StatusOK)
}
