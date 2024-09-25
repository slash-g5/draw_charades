package dto

import "encoding/json"

type JsonBinaryMarshaler struct {
}

func (jbm *JsonBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return json.Marshal(jbm)
}
