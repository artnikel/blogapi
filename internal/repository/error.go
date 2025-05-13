// Package repository error.go contains custom errors
package repository

import "fmt"

// ErrNil means that u've given nil entity for a create method
var ErrNil = fmt.Errorf("entity that u've given is nil")

// ErrExist means that u've given username that already exist
var ErrExist = fmt.Errorf("such username already exist")
