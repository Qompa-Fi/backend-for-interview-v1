package main

import "encoding/json"

func mustJSONEncode[T any](v T) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}
