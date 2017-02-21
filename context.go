package zen

import (
	"encoding/asn1"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type (
	// Context warps request and response writer
	Context struct {
		req *http.Request
		rw  http.ResponseWriter
	}
)

// JSON : write json data to http response writer
func (z *Context) JSON(code int, i interface{}) (err error) {
	// write http status code
	z.rw.Header().Add("content-type", "application/json")
	z.rw.WriteHeader(code)

	// Encode json data to rw
	err = json.NewEncoder(z.rw).Encode(i)

	//return
	return
}

// XML : write xml data to http response writer
func (z *Context) XML(code int, i interface{}) (err error) {
	// write http status code
	z.rw.Header().Add("content-type", "application/xml")
	z.rw.WriteHeader(code)

	// Encode xml data to rw
	err = xml.NewEncoder(z.rw).Encode(i)

	//return
	return
}

// ASN1 : write asn1 data to http response writer
func (z *Context) ASN1(code int, i interface{}) (err error) {
	// write http status code
	z.rw.Header().Add("content-type", "application/asn1")
	z.rw.WriteHeader(code)

	// Encode asn1 data to rw
	bts, err := asn1.Marshal(i)
	if err != nil {
		return
	}
	//return
	_, err = z.rw.Write(bts)
	return
}
