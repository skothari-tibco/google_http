package google_http

import (
	"github.com/project-flogo/core/data/coerce"
)

type Output struct {
	PathParams  map[string]string `md:"pathParams"`
	QueryParams map[string]string `md:"queryParams"`
	Headers     map[string]string `md:"headers"`
	Content     interface{}       `md:"content"`
}
type Reply struct {
	Data   interface{} `md:"data"`
	Status int         `md:"status"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"pathParams":  o.PathParams,
		"queryParams": o.QueryParams,
		"headers":     o.Headers,
		"content":     o.Content,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.PathParams, err = coerce.ToParams(values["pathParams"])
	if err != nil {
		return err
	}
	o.QueryParams, err = coerce.ToParams(values["queryParams"])
	if err != nil {
		return err
	}
	o.Headers, err = coerce.ToParams(values["headers"])
	if err != nil {
		return err
	}
	o.Content = values["content"]

	return nil

}

func (r *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"status": r.Status,
		"data":   r.Data,
	}
}

func (r *Reply) FromMap(values map[string]interface{}) error {

	var err error
	r.Status, err = coerce.ToInt(values["status"])
	if err != nil {
		return err
	}
	r.Data, _ = values["data"]

	return nil
}
