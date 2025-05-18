// Package constants - all project constants
package constants

import "time"

const (
	// ServerTimeout — the maximum duration for the server to wait for active connections to finish during shutdown
	ServerTimeout = 10 * time.Second

	// AccessTokenExpiration — the lifespan of the Access Token before it expires
	AccessTokenExpiration = 15 * time.Minute

	// RefreshTokenExpiration — the lifespan of the Refresh Token before it expires
	RefreshTokenExpiration = 72 * time.Hour

	// BcryptCost — the hashing cost (complexity) for bcrypt when encrypting passwords
	BcryptCost = 14
)
