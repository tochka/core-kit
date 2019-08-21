package codec

import (
	"mime"
	"strings"
)

var registeredCodecs = make(map[string]map[string]Codec)

// RegisterCodec registers the provided Codec for use with all clients and
// servers.
//
// The Codec will be stored and looked up by result of its Name() method, which
// should match the content-type of the encoding handled by the Codec.  This
// is case-insensitive, and is stored and looked up as lowercase.  If the
// result of calling Name() is an empty string, Register will panic. See
// Content-Type on
// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests for
// more details.
//
// NOTE: this function must only be called during initialization time (i.e. in
// an init() function), and is not thread-safe.  If multiple Compressors are
// registered with the same name, the one registered last will take effect.
func Register(codec Codec) {
	if codec == nil {
		panic("cannot register a nil Codec")
	}
	contentType := strings.ToLower(codec.Name())
	if contentType == "" {
		panic("cannot register Codec with empty string result for Name()")
	}
	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		panic("codec type incorrect: " + err.Error())
	}

	indx := strings.Index(mimeType, "/")

	subType, ok := registeredCodecs[mimeType[:indx]]
	if !ok {
		subType = make(map[string]Codec)
		registeredCodecs[mimeType[:indx]] = subType
	}
	subType[mimeType[indx+1:]] = codec
}

// GetCodec gets a registered Codec by content-subtype, or nil if no Codec is
// registered for the content-subtype.
//
// The content-subtype is expected to be lowercase.
func Get(contentType string) Codec {
	indx := strings.Index(contentType, "/")
	if indx == -1 {
		return nil
	}
	subType := registeredCodecs[contentType[:indx]]
	if subType == nil {
		return nil
	}
	if contentType[indx+1:] == "*" {
		for k := range subType {
			return subType[k]
		}
	}
	return subType[contentType[indx+1:]]
}

//
// Codec is interface for marshal/unmarshal
//
type Codec interface {
	Marshal(obj interface{}) (body []byte, err error)
	Unmarshal(data []byte, obj interface{}) error
	Name() string
}
