package relay

import "net/http"

//IsWebSocketRequest returns true if a http.Request object is based on
//websocket standards
func IsWebSocketRequest(r *http.Request) bool {
	var _ interface{}
	_, hasUpgrade := r.Header["Upgrade"]
	_, hasSec := r.Header["Sec-Websocket-Version"]
	_, hasExt := r.Header["Sec-Websocket-Extensions"]
	_, hasKey := r.Header["Sec-Websocket-Key"]
	return hasUpgrade && hasSec && hasExt && hasKey
}
