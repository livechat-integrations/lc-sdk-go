package internal

import "encoding/json"

func UnmarshalOptionalRawField(source json.RawMessage, target interface{}) error {
	if source != nil {
		return json.Unmarshal(source, target)
	}
	return nil
}
