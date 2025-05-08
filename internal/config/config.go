package config

type Config struct {
	BlogPostgresPath   string `env:"BLOG_POSTGRES_PATH"`
	BlogTokenSignature string `env:"BLOG_TOKEN_SIGNATURE"`
}
