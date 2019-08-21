package json

import (
	"encoding/base64"
	"encoding/json"

	"github.com/tochka/core-kit/codec"
	"github.com/tochka/core-kit/errors"
)

// Name is the name registered for the JSON coded.
const Name = "application/json"

func init() {
	codec.Register(NewCodec())
}

func NewCodec() Codec {
	return Codec{}
}

type Codec struct{}

func (Codec) Marshal(obj interface{}) (body []byte, err error) {
	errCtx := errors.Context("component", "codec", "type", "json", "method", "marshal", "data", obj)
	defer errors.Defer(&err, errCtx)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (Codec) Unmarshal(data []byte, obj interface{}) error {
	if err := json.Unmarshal(data, obj); err != nil {
		return errors.Wrap(err, "component", "codec", "type", "strict_json", "method", "unmarshal",
			"data", base64.StdEncoding.EncodeToString(data), "receiver", obj)
	}
	return nil
}

func (Codec) Name() string {
	return Name
}
