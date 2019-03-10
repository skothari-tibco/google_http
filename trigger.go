package google_http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/project-flogo/core/support/log"

	"github.com/project-flogo/core/trigger"
)

var triggerMd = trigger.NewMetadata(&Output{}, &Reply{})
var singleton *GoogleHttpTrigger

func init() {
	trigger.Register(&GoogleHttpTrigger{}, &GoogleHttpFactory{})
}

// LambdaFactory AWS Lambda Trigger factory
type GoogleHttpFactory struct {
}

//New Creates a new trigger instance for a given id
func (t *GoogleHttpFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	singleton = &GoogleHttpTrigger{}
	return singleton, nil

}

// Metadata implements trigger.Trigger.Metadata
func (t *GoogleHttpFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

// LambdaTrigger AWS Lambda trigger struct
type GoogleHttpTrigger struct {
	id       string
	log      log.Logger
	handlers []trigger.Handler
}

func (t *GoogleHttpTrigger) Initialize(ctx trigger.InitContext) error {
	t.id = "Google"
	t.log = ctx.Logger()
	t.handlers = ctx.GetHandlers()
	return nil
}

// Invoke starts the trigger and invokes the action registered in the handler
func Invoke(w http.ResponseWriter, r *http.Request) {

	out := &Output{}

	out.PathParams = make(map[string]string)
	/*
		for _, param := range ps {
			out.PathParams[param.Key] = param.Value
		}*/

	queryValues := r.URL.Query()
	out.QueryParams = make(map[string]string, len(queryValues))
	out.Headers = make(map[string]string, len(r.Header))

	for key, value := range r.Header {
		out.Headers[key] = strings.Join(value, ",")
	}

	for key, value := range queryValues {
		out.QueryParams[key] = strings.Join(value, ",")
	}

	// Check the HTTP Header Content-Type
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/x-www-form-urlencoded":
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		m, err := url.ParseQuery(s)
		content := make(map[string]interface{}, 0)
		if err != nil {
			singleton.log.Errorf("Error while parsing query string: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		for key, val := range m {
			if len(val) == 1 {
				content[key] = val[0]
			} else {
				content[key] = val[0]
			}
		}

		out.Content = content
	case "application/json":
		var content interface{}
		err := json.NewDecoder(r.Body).Decode(&content)
		if err != nil {
			switch {
			case err == io.EOF:
				// empty body
				//todo should handler say if content is expected?
			case err != nil:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		out.Content = content
	default:
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		out.Content = string(b)
	}

	results, err := singleton.handlers[0].Handle(context.Background(), out)

	reply := &Reply{}
	reply.FromMap(results)

	if err != nil {
		singleton.log.Debugf("Error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if reply.Data != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if reply.Status == 0 {
			reply.Status = 200
		}
		w.WriteHeader(reply.Status)
		if err := json.NewEncoder(w).Encode(reply.Data); err != nil {
			log.Error(err)
		}
		return
	}

	if reply.Status > 0 {
		w.WriteHeader(reply.Status)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (t *GoogleHttpTrigger) Start() error {
	return nil
}

// Stop implements util.Managed.Stop
func (t *GoogleHttpTrigger) Stop() error {
	return nil
}
