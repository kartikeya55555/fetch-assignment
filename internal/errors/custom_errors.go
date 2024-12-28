package errors

import "errors"

var (
	ErrNotValidTimeFormat  = errors.New("time of purchase is not valid")
	ErrReceiptNotExist     = errors.New("receiptId doesn't exist")
	ErrNotValidTotalFormat = errors.New("total format is not valid")
	ErrNotValidDateFormat  = errors.New("date format is not valid")
)
