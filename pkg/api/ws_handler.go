package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

type wsRequest struct {
	ID     int                    `json:"int"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type wsResponse struct {
	ID       int         `json:"int"`
	Error    string      `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
}

func wsJSON(session sockjs.Session, input interface{}) {
	data, err := json.Marshal(input)
	if err != nil {
		return
	}
	session.Send(string(data))
}

func (a *API) ws(session sockjs.Session) {
	for {
		msg, err := session.Recv()
		if err != nil {
			a.Log.Warn("SockJS session error", "error", err)
			break
		}

		var req *wsRequest
		if err := json.Unmarshal([]byte(msg), &req); err != nil {
			a.Log.Warn("Error while unmarshalling a JSON input", "error", err)
			wsJSON(session, &wsResponse{
				ID:    0,
				Error: err.Error(),
			})
			continue
		}

		switch req.Method {
		case "request":
			// Map params from request to actual request
			var (
				method  string
				path    string
				body    string
				headers http.Header
			)
			if mi, ok := req.Params["method"]; ok {
				if mv, ok := mi.(string); ok {
					method = mv
				}
			}
			if pi, ok := req.Params["path"]; ok {
				if pv, ok := pi.(string); ok {
					path = pv
				}
			}
			if bi, ok := req.Params["body"]; ok {
				if bv, ok := bi.(string); ok {
					body = bv
				}
			}
			if hi, ok := req.Params["headers"]; ok {
				// map[string]interface{} -> map[string]map[string]interface{}
				if hv, ok := hi.(map[string]interface{}); ok {
					for kv, li := range hv {
						// map[string]map[string]interface{} -> map[string]map[string][]interface{}
						if lv, ok := li.([]interface{}); ok {
							for _, ii := range lv {
								// map[string]map[string][]interface{} -> map[string]map[string][]string
								if iv, ok := ii.(string); ok {
									headers.Add(kv, iv)
								}
							}
						}
					}
				}
			}

			if method == "" || path == "" {
				a.Log.Warn("No method or path in a WS request", "id", session.ID())
				wsJSON(session, &wsResponse{
					ID:    req.ID,
					Error: "Method or path empty",
				})
				continue
			}

			// Normalize the path into an URL
			if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
				if path[0] != '/' {
					path = "/" + path
				}

				path = a.Config.URL + path
			}

			// Generate a new request
			re, err := http.NewRequest(method, path, strings.NewReader(body))
			if err != nil {
				a.Log.Warn("Unable to create a new HTTP request", "error", err, "id", session.ID())
				wsJSON(session, &wsResponse{
					ID:    req.ID,
					Error: err.Error(),
				})
			}
			re.Header = headers

			// Prepare a new response capturer
			rw := httptest.NewRecorder()

			// Run the request
			a.Router.ServeHTTP(rw, re)

			// Generate a response
			resp := &wsResponse{
				ID: req.ID,
				Response: map[string]interface{}{
					"code":    rw.Code,
					"headers": rw.Header(),
				},
			}
			if rw.Body != nil {
				resp.Response.(map[string]interface{})["body"] = rw.Body.String()
			}

			// Write a response
			wsJSON(session, resp)
		}
	}
}
