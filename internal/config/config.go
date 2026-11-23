// Package config loads all application configuration from environment variables.
// Call InitConfig once at startup and pass the returned *Config to all layers via dependency injection.
package config

import "time"

// LoggingConfig controls log verbosity and output paths.
// EnvType controls the log level: LOCAL-DEV → Debug, all others → Info.
type LoggingConfig struct {
	EnvType          string // LOCAL-DEV | STAGING | PRODUCTION
	LogFilePath      string
	ErrorLogFilePath string
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Debug         bool
	ServerAddress string
	ServerPort    string
	EnvPath       string
}

// SiteConfig holds platform identity fields injected into emails and responses.
type SiteConfig struct {
	SiteName   string
	SiteDomain string
	SiteEmail  string
	BaseURL    string // e.g. https://myapp.com
	BaseAPIURL string // e.g. https://api.myapp.com
}

// DBConfig holds database connection settings.
// Leave DBURL empty to use SQLite (development). Set it for PostgreSQL (production).
type DBConfig struct {
	DBURL        string
	DBDRIVER     string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
	MaxLifeTime  time.Duration
}

// SMTPProviderConfig holds credentials for one SMTP sender.
type SMTPProviderConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // display address, e.g. "MyApp <info@myapp.com>"
	ReplyTo  string
}

// EmailConfig holds named SMTP providers. Add a Backup field to enable fallback sending.
type EmailConfig struct {
	Primary SMTPProviderConfig
	// Backup SMTPProviderConfig  // uncomment to add a fallback provider
}

// CloudinaryConfig holds Cloudinary storage credentials.
type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	URL       string
}

// StorageConfig holds file storage settings.
type StorageConfig struct {
	CloudinaryConfig CloudinaryConfig
}

// JWTConfig holds signing key and token lifetime settings.
type JWTConfig struct {
	JWTSecretKey  string
	Issuer        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// TelegramConfig holds credentials for the Telegram alert bot.
// Leave empty to disable Telegram alerts.
type TelegramConfig struct {
	Token  string
	ChatID string
}

// Config is the top-level configuration object. Injected into the service layer at startup.
type Config struct {
	SiteConfig    SiteConfig
	ServerConfig  ServerConfig
	DBConfig      DBConfig
	EmailConfig   EmailConfig
	JWTConfig     JWTConfig
	StorageConfig StorageConfig
	TelegramConfig TelegramConfig
	LoggingConfig  LoggingConfig
}

// InitConfig loads config from the .env file at envFilePath and returns a populated *Config.
func InitConfig(envFilePath string) *Config {
	loadEnvFile(envFilePath)

	return &Config{
		ServerConfig: ServerConfig{
			Debug:         getEnvBool("DEBUG", true),
			ServerAddress: getEnv("SERVER_ADDRESS", "0.0.0.0"),
			ServerPort:    getEnv("SERVER_PORT", "8080"),
			EnvPath:       envFilePath,
		},
		SiteConfig: SiteConfig{
			SiteName:   getEnv("SITE_NAME", "MyApp"),
			SiteDomain: getEnv("SITE_DOMAIN", "myapp.com"),
			SiteEmail:  getEnv("SITE_EMAIL", "info@myapp.com"),
			BaseURL:    getEnv("BASE_URL", "http://localhost:3000"),
			BaseAPIURL: getEnv("BASE_API_URL", "http://localhost:8080"),
		},
		LoggingConfig: LoggingConfig{
			EnvType:          getEnv("ENV_TYPE", "LOCAL-DEV"),
			LogFilePath:      getEnv("LOG_FILE_PATH", "./logs/app.log"),
			ErrorLogFilePath: getEnv("ERROR_LOG_FILE_PATH", "./logs/errors.log"),
		},
		DBConfig: DBConfig{
			DBURL:        getEnv("DBURL", ""),
			DBDRIVER:     getEnv("DBDRIVER", "sqlite"),
			MaxOpenConns: 50,
			MaxIdleConns: 10,
			MaxIdleTime:  mustParseDuration("15m"),
			MaxLifeTime:  mustParseDuration("1h"),
		},
		JWTConfig: JWTConfig{
			JWTSecretKey:  getEnv("JWT_SECRET_KEY", ""),
			Issuer:        getEnv("JWT_ISSUER", "myapp"),
			AccessExpiry:  time.Hour * 24,      // 24 hours — adjust per project
			RefreshExpiry: time.Hour * 24 * 7,  // 7 days
		},
		EmailConfig: EmailConfig{
			Primary: SMTPProviderConfig{
				Host:     getEnv("PRIMARY_EMAIL_HOST", ""),
				Port:     getEnvInt("PRIMARY_EMAIL_PORT", 587),
				Username: getEnv("PRIMARY_EMAIL_USERNAME", ""),
				Password: getEnv("PRIMARY_EMAIL_PASSWORD", ""),
				From:     getEnv("PRIMARY_EMAIL_FROM", ""),
				ReplyTo:  getEnv("PRIMARY_EMAIL_REPLY_TO", ""),
			},
		},
		StorageConfig: StorageConfig{
			CloudinaryConfig: CloudinaryConfig{
				CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
				APIKey:    getEnv("CLOUDINARY_API_KEY", ""),
				APISecret: getEnv("CLOUDINARY_API_SECRET", ""),
				URL:       getEnv("CLOUDINARY_URL", ""),
			},
		},
		TelegramConfig: TelegramConfig{
			Token:  getEnv("TELEGRAM_BOT_TOKEN", ""),
			ChatID: getEnv("TELEGRAM_CHAT_ID", ""),
		},
	}
}
