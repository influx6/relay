package relay

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/influx6/flux"
)

// BuildMatchesMethod returns RHandler which matches the methods of a request to a list of accepted methods
// to decided wether to run a RHandler function
func BuildMatchesMethod(mo []string, fx RHandler) RHandler {
	return func(w http.ResponseWriter, r *http.Request, c Collector) {
		MatchesMethod(mo, w, r, c, fx)
	}
}

// MatchesMethod provides a factory function for matching sets of methods against a reqests method
func MatchesMethod(mo []string, w http.ResponseWriter, r *http.Request, c Collector, fx RHandler) {
	method := strings.ToLower(r.Method)
	//ok we have no method restrictions or we have the method,so we can continue else ignore
	if len(mo) == 0 || HasMethod(mo, method) {
		fx(w, r, c)
	}
}

var multispaces = regexp.MustCompile(`\s+`)

// GetMethods turns a string of space separated methods into a list
func GetMethods(m string) []string {
	var methods []string

	if m != "" {
		cln := multispaces.ReplaceAllString(m, " ")
		methods = multispaces.Split(cln, -1)
	}

	return methods
}

// HasMethod returns true/false if a method is found in a list of method names
func HasMethod(mo []string, m string) bool {
	var found bool
	m = strings.ToLower(m)

	for _, mi := range mo {
		if m != mi {
			continue
		}
		found = true
		break
	}

	return found
}

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

func setUpHeadings(r *Context) {
	agent, ok := r.Req.Header["User-Agent"]

	if ok {
		ag := strings.Join(agent, ";")
		msie := strings.Index(ag, ";MSIE")
		trident := strings.Index(ag, "Trident/")

		if msie != -1 || trident != -1 {
			// r.Res.Header().Set("X-XSS-Protection", "0")
		}
	}

	origin, ok := r.Req.Header["Origin"]

	if ok {
		r.Res.Header().Set("Access-Control-Allow-Credentials", "true")
		r.Res.Header().Set("Access-Control-Allow-Origin", strings.Join(origin, ";"))
	} else {
		r.Res.Header().Set("Access-Control-Allow-Origin", "*")
	}
}

// ErrNoBody is returned when the request has no body
var ErrNoBody = errors.New("Http Request Has no body")

func loadData(r *Context) (*Message, error) {
	msg := Message{}
	msg.Method = r.Req.Method

	msg.Queries = r.Req.URL.Query()

	content, ok := r.Req.Header["Content-Type"]

	if ok {
		muxcontent := strings.ToLower(strings.Join(content, ";"))

		if strings.Index(muxcontent, "application/x-www-form-urlencode") != -1 {
			if err := r.Req.ParseForm(); err != nil {
				return nil, err
			}

			msg.MessageType = "form"
			msg.Method = r.Req.Method
			msg.Form = r.Req.Form
			msg.PostForm = r.Req.PostForm

			return &msg, nil
		}

		if strings.Index(muxcontent, "multipart/form-data") != -1 {
			if err := r.Req.ParseMultipartForm(32 << 20); err != nil {
				return nil, err
			}

			msg.MessageType = "multipart"
			msg.Multipart = r.Req.MultipartForm
			return &msg, nil
		}
	}

	if r.Req.Body == nil {
		if err := r.Req.ParseForm(); err != nil {
			return nil, err
		}
		msg.Form = r.Req.Form
		msg.PostForm = r.Req.PostForm
		//set the type to basic,meaning no body type just requsts and query params
		msg.MessageType = "basic"
		return &msg, nil
	}

	data := make([]byte, r.Req.ContentLength)
	_, err := r.Req.Body.Read(data)

	if err != nil && err != io.EOF {
		return nil, err
	}

	msg.MessageType = "body"
	msg.Payload = data

	return &msg, nil
}

func expect(t *testing.T, v, m interface{}) {
	if v != m {
		flux.FatalFailed(t, "Value %+v and %+v are not a match", v, m)
		return
	}
	flux.LogPassed(t, "Value %+v and %+v are a match", v, m)
}
