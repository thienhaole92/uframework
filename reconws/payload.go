package reconws

import (
	"encoding/json"
	"errors"
)

var ErrPayloadEmpty = errors.New("payload empty")

type Payload []byte

func (p Payload) Unpack(v any) error {
	if len(p) == 0 {
		return ErrPayloadEmpty
	}

	return json.Unmarshal(p, v)
}

func (p Payload) ToMap() map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal(p, &out)

	return out
}
