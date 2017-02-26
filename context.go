package zen

import (
	"bytes"
	"encoding/asn1"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	inputTagName = "form"
	validTagName = "valid"
	validMsgName = "msg"
)

//commonly used mime-types
const (
	applicationJSON = "application/json"
	applicationXML  = "application/xml"
	textXML         = "text/xml"
	applicationASN1 = "application/asn1"

	contentType = "Content-Type"
)

type (
	// Context warps request and response writer
	Context struct {
		Req    *http.Request
		rw     http.ResponseWriter
		params map[string]string
		parsed bool
	}
)

func (s *Server) getContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := s.contextPool.Get().(*Context)
	c.Req = req
	c.rw = rw
	return c
}

func (s *Server) putBackContext(c *Context) {
	c.params = nil
	c.parsed = false
	c.Req = nil
	c.rw = nil

	s.contextPool.Put(c)
}

// ParseInput will parse request's form and
func (c *Context) ParseInput() error {
	if err := c.Req.ParseForm(); err != nil {
		return err
	}
	return nil
}

// Form return request form
func (c *Context) Form() url.Values {
	return c.Req.Form
}

// Params return request params
func (c *Context) Params() map[string]string {
	return c.params
}

// RequestHeader return request's header
func (c *Context) RequestHeader() http.Header {
	return c.Req.Header
}

// ParseValidateForm will parse request's form and map into a interface{} value
func (c *Context) ParseValidateForm(input interface{}) error {
	if err := c.Req.ParseForm(); err != nil {
		return err
	}
	return c.parseValidateForm(input)
}

func (c *Context) parseValidateForm(input interface{}) error {
	inputValue := reflect.ValueOf(input).Elem()
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		tag := inputType.Field(i).Tag
		formName := tag.Get(inputTagName)
		validate := tag.Get(validTagName)
		validateMsg := tag.Get(validMsgName)
		field := inputValue.Field(i)
		formValue := c.Req.Form.Get(formName)

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

	}
	return nil
}

func valid(s string, validate, msg string) error {
	if validate == "" {
		return nil
	}
	rxp, err := regexp.Compile(validate)
	if err != nil {
		return err
	}

	if !rxp.MatchString(s) {
		return errors.New(msg)
	}

	return nil
}

// JSON : write json data to http response writer, with status code 200
func (c *Context) JSON(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationJSON)

	// Encode json data to rw
	err = json.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// XML : write xml data to http response writer, with status code 200
func (c *Context) XML(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationXML)

	// Encode xml data to rw
	err = xml.NewEncoder(c.rw).Encode(i)

	//return
	return
}

// ASN1 : write asn1 data to http response writer, with status code 200
func (c *Context) ASN1(i interface{}) (err error) {
	// write http status code
	c.WriteHeader(contentType, applicationASN1)

	// Encode asn1 data to rw
	bts, err := asn1.Marshal(i)
	if err != nil {
		return
	}
	//return
	_, err = c.rw.Write(bts)
	return
}

// WriteStatus set response's status code
func (c *Context) WriteStatus(code int) {
	c.rw.WriteHeader(code)
}

// WriteHeader set response header
func (c *Context) WriteHeader(k, v string) {
	c.rw.Header().Add(k, v)
}

// Raw write raw bytes
func (c *Context) Raw(b []byte) {
	io.Copy(c.rw, bytes.NewReader(b))
}

// RawStr write raw string
func (c *Context) RawStr(s string) {
	io.Copy(c.rw, strings.NewReader(s))
}
