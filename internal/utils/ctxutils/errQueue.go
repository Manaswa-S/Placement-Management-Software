package ctxutils

import (
	"go.mod/cmd/server/errHandler"
	"go.mod/internal/dto"
)


func NewError(report *dto.ErrorData) {

	errHandler.ErrorsChan <- report
}