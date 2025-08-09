package handlers

import (
	"net/http"
	"net/http/httputil"
)

// Echo is an http.HandlerFunc that echoes the request back to the caller.
func Echo(rw http.ResponseWriter, req *http.Request) {
	reqBs, err := httputil.DumpRequest(req, true)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(reqBs)
}
