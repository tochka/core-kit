package json

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/tochka/core-kit/errors"
)

func NewStrictCodec() StrictCode {
	return StrictCode{}
}

type StrictCode struct{}

func (StrictCode) Marshal(obj interface{}) (body []byte, err error) {
	errCtx := errors.Context("component", "codec", "type", "strict_json", "method", "marshal", "data", obj)
	defer errors.Defer(&err, errCtx)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (StrictCode) Unmarshal(data []byte, obj interface{}) (err error) {
	errCtx := errors.Context("component", "codec", "type", "strict_json", "method", "unmarshal",
		"data", base64.StdEncoding.EncodeToString(data), "receiver", obj)
	defer errors.Defer(&err, errCtx)

	if !json.Valid(data) {
		return errors.Wrap(fmt.Errorf("validate JSON"))
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(obj); err != nil {
		return err
	}
	return nil
}

func (StrictCode) Name() string {
	return Name
}
