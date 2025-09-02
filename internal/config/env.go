package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port           string
	MongoURI       string
	DBName         string
	JWTSecret      string
	EnableRegister bool
}

var Cfg AppConfig

func Load() {
	_ = godotenv.Load()
	Cfg = AppConfig{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("MONGO_DB", "task_manager"),
		JWTSecret:      getEnv("JWT_SECRET", "devsecret"),
		EnableRegister: getEnv("ENABLE_REGISTER", "false") == "true",
	}
	log.Printf("[config] loaded. DB=%s Port=%s", Cfg.DBName, Cfg.Port)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
