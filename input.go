package zen

import (
	"net/url"
)

type (
	// Input is a in abstract struct that unmarshal web forms into struct, then valid the input base on field's tag
	Input struct {
	}
)

func (i *Input) unmarshal(form url.Values) {

}

func (i *Input) valid() error {
	return nil
}
