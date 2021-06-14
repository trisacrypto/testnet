/*
 * Package jsonpb uses protojson (not the deprecated jsonpb module) to set defaults for
 * marshaling and unmarshaling rVASP protobuf messages to and from JSON format.
 */
package jsonpb

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshaler = &protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		AllowPartial:    true,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}

	unmarshaler = &protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
)

func Marshal(m proto.Message) ([]byte, error) {
	return marshaler.Marshal(m)
}

func Unmarshal(data []byte, m proto.Message) error {
	return unmarshaler.Unmarshal(data, m)
}

func UnmarshalString(data string, m proto.Message) error {
	return unmarshaler.Unmarshal([]byte(data), m)
}
