package repository

import "fmt"

var ErrNil = fmt.Errorf("entity that u've given is nil")

var ErrExist = fmt.Errorf("such username already exist")
