package main;

import (
	"github.com/buger/jsonparser"
)

type JsonObj struct {
	Data []byte
}

func (o JsonObj) GetStr(keys ...string) string {
	str, _ := jsonparser.GetString(o.Data, keys...)
	return str
}

func (o JsonObj) GetInt(keys ...string) int64 {
	i, _ := jsonparser.GetInt(o.Data, keys...)
	return i
}
