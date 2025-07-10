package utils

import "errors"

var ErrAllPaymentsFailing = errors.New("all payment processors failing")
var ErrPaymentAlreadyExists = errors.New("payment already exists")
