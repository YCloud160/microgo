package errors

import (
	"encoding/json"
)

type Error struct {
	Id   string `json:"id"`
	Code int32  `json:"code"`
	Desc string `json:"desc"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func New(id, desc string, code int32) error {
	return &Error{
		Id:   id,
		Code: code,
		Desc: desc,
	}
}

func ParseError(err string) *Error {
	e := new(Error)
	jerr := json.Unmarshal([]byte(err), e)
	if jerr != nil {
		e.Code = 9999
		e.Desc = err
	}
	return e
}
