package api_err

import "fmt"

type ApiErr struct {
	Typ        ApiErrTyp `json:"typ"`
	Msg        string `json:"msg"`
	HttpStatus int    `json:"-"`
}

func (e *ApiErr) Status() int {
	if e.HttpStatus == 0 {
		return 400
	}
	return e.HttpStatus
}

func (e *ApiErr) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Status(), e.Typ, e.Msg)
}

func NewApiErr(typ ApiErrTyp, msgFmt string, args ...any) *ApiErr {
	return &ApiErr{Typ: typ, Msg: fmt.Sprintf(msgFmt, args...)}
}

func NewApiErrS(status int, typ ApiErrTyp, msgFmt string, args ...any) *ApiErr {
	return &ApiErr{Typ: typ, Msg: fmt.Sprintf(msgFmt, args...), HttpStatus: status}
}
