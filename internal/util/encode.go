package util

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

func Unmarshal[Tval any](data string) (v Tval, err error) {
	rawData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(rawData)
	gobdec := gob.NewDecoder(buffer)

	err = gobdec.Decode(&v)
	if err != nil {
		return
	}

	return
}

func Marshal[Tval any](v Tval) (data string, err error) {
	var buffer bytes.Buffer
	gobenc := gob.NewEncoder(&buffer)

	err = gobenc.Encode(v)
	if err != nil {
		return
	}

	data = base64.StdEncoding.EncodeToString(buffer.Bytes())

	return
}
