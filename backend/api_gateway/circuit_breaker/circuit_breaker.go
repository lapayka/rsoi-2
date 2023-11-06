package circuit_breaker

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lapayka/rsoi-2/Common/Logger"
)

type state int

const (
	Open     state = 0
	HalfOpen       = 1
	Closed         = 2
)

type CircuitBreaker struct {
	n    int
	Host string
	st   state
}

const N int = 3

func GetDefaultClient() *http.Client {
	client := http.DefaultClient

	return client
}

func (cb *CircuitBreaker) SetTimer() {
	timer := time.NewTimer(9 * time.Second)
	go func() {
		<-timer.C
		cb.CheckServiceState()
	}()
}

func (cb *CircuitBreaker) CheckServiceState() {
	if cb.st == Closed {
		fmt.Printf("Sending healthchek reuest to %s\n", cb.Host)
		resp, err := GetDefaultClient().Get(cb.Host + "/manage/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			cb.st = Open
			cb.n = 0
			fmt.Printf("Service %s is alive", cb.Host)
		} else {
			fmt.Println(err)
			cb.SetTimer()
		}
	}
}

func (cb *CircuitBreaker) ProxyQuery(w http.ResponseWriter, r *http.Request) {
	if cb.st == Closed {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req, _ := http.NewRequest(r.Method, cb.Host+r.URL.String(), r.Body)
	req.Header = r.Header
	response, err := GetDefaultClient().Do(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		cb.n = cb.n + 1
		if cb.n > N || cb.st == HalfOpen {
			cb.st = Closed
			cb.SetTimer()
		}
	} else {
		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(response.StatusCode)
		io.Copy(w, response.Body)

		if cb.st == HalfOpen {
			cb.st = Open
		}
	}

}

func (cb *CircuitBreaker) SendQuery(r *http.Request) (*http.Response, error) {
	if cb.st == Closed {
		return nil, fmt.Errorf("Service %s closed", cb.Host)
	}

	req, _ := http.NewRequest(r.Method, cb.Host+r.URL.String(), r.Body)

	req.Header = r.Header
	response, err := GetDefaultClient().Do(req)

	if err != nil {
		response = &http.Response{StatusCode: http.StatusInternalServerError}
		cb.n = cb.n + 1
		if cb.n > N || cb.st == HalfOpen {
			cb.st = Closed
			cb.SetTimer()
		}
	} else {
		if cb.st == HalfOpen {
			cb.st = Open
		}
	}

	return response, err
}
