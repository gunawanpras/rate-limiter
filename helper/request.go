package helper

import "net/http"

type (
	Request interface {
		GetIp(r *http.Request) string
	}

	reqHelper struct{}
)

func (h *reqHelper) GetIp(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// If the X-Forwarded-For header is not present, use the RemoteAddr field
	return r.RemoteAddr
}

var ReqHelper Request = &reqHelper{}
