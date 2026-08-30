package database

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

func CompressJSON(buf *bytes.Buffer, obj any) error {
	gz, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err := json.NewEncoder(gz).Encode(obj); err != nil {
		return err
	}

	return gz.Close()
}

func DecompressJSON[T any](b []byte) (T, error) {
	var res T
	reader := bytes.NewReader(b)

	gz, err := gzip.NewReader(reader)
	if err != nil {
		return res, err
	}

	if err := json.NewDecoder(gz).Decode(&res); err != nil {
		return res, err
	}

	return res, gz.Close()
}
