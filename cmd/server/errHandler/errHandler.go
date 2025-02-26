package errHandler

import "go.mod/internal/dto"


var ErrorsChan = make(chan *dto.ErrorData, 50)

