// Package config represents structure Config
package config

// Config is a structure of environment variables
type Config struct {
	BlogPostgresPath     string `env:"BLOG_POSTGRES_PATH"`
	BlogTokenSignature   string `env:"BLOG_TOKEN_SIGNATURE"`
	BlogServerPort       string `env:"BLOG_SERVER_PORT"`
	BlogPostgresDB       string `env:"BLOG_POSTGRES_DB"`
	BlogPostgresUser     string `env:"BLOG_POSTGRES_USER"`
	BlogPostgresPassword string `env:"BLOG_POSTGRES_PASSWORD"`
}
