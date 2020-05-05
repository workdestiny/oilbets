package entity

import "errors"

var (
	// ErrNotFound ใช้ตรวจสอบ err ใน repo
	ErrNotFound = errors.New("Not Found")
)
