package zen

import (
	"encoding/asn1"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
)

const (
	inputTagName = "form"
	validTagName = "valid"
	validMsgName = "msg"
)

type (
	// Context warps request and response writer
	Context struct {
		req *http.Request
		rw  http.ResponseWriter
	}
)

var (
	errUnsupportedFromType = errors.New("unsupported form type")
)

// ParseValidForm will parse request's form and map into a interface{} value
func (c *Context) ParseValidForm(input interface{}) error {
	if err := c.req.ParseForm(); err != nil {
		return err
	}
	return parseValidForm(input, c.req.Form)
}

func parseValidForm(input interface{}, form url.Values) error {
	inputValue := reflect.ValueOf(input)
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		tag := inputType.Field(i).Tag
		formName := tag.Get(inputTagName)
		validate := tag.Get(validTagName)
		validateMsg := tag.Get(validMsgName)
		field := inputValue.Field(i)
		formValue := form.Get(formName)

		// validate form with regex
		if err := valid(formValue, validate, validateMsg); err != nil {
			return err
		}
		// scan form string value into field
		if err := scan(field, formValue); err != nil {

			return err
		}

	}
	return nil
}

func scan(v reflect.Value, s string) error {

	if !v.CanSet() {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(x)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)

	default:
		return errUnsupportedFromType
	}
	return nil
}

func valid(s string, validate, msg string) error {
	rxp, err := regexp.Compile(validate)
	if err != nil {
		return err
	}

	if !rxp.MatchString(s) {
		return errors.New(msg)
	}

	return nil
}

// JSON : write json data to http response writer
func (c *Context) JSON(code int, i interface{}) (err error) {
	// write http status code
	c.rw.Header().Add("content-type", "application/json")
	c.rw.WriteHeader(code)

	// Encode json data to rw
	err = json.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// XML : write xml data to http response writer
func (c *Context) XML(code int, i interface{}) (err error) {
	// write http status code
	c.rw.Header().Add("content-type", "application/xml")
	c.rw.WriteHeader(code)

	// Encode xml data to rw
	err = xml.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// ASN1 : write asn1 data to http response writer
func (c *Context) ASN1(code int, i interface{}) (err error) {
	// write http status code
	c.rw.Header().Add("content-type", "application/asn1")
	c.rw.WriteHeader(code)

	// Encode asn1 data to rw
	bts, err := asn1.Marshal(i)
	if err != nil {
		return
	}
	//return
	_, err = c.rw.Write(bts)
	return
}
