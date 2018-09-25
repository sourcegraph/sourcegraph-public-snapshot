package db

import "github.com/lib/pq"

func isPQErrorCode(err error, code string) bool {
	errp, ok := err.(*pq.Error)
	return ok && errp.Code == pq.ErrorCode(code)
}

func isPQErrorUniqueViolation(err error) bool {
	return isPQErrorCode(err, "23505")
}
