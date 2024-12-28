package config

import (
	"os"
)

// Config holds all configuration values for the app
type Config struct {
	ServiceType  string // "api" or "worker"
	Port         string // defaults to "8080"
	AWSRegion    string
	AWSEndpoint  string
	SQSQueueName string
	SNSTopicName string
}

// LoadConfig loads environment variables into a Config struct.
func LoadConfig() *Config {
	return &Config{
		ServiceType:  getEnv("SERVICE_TYPE", "api"),
		Port:         getEnv("PORT", "8080"),
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint:  getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		SQSQueueName: getEnv("SQS_QUEUE_NAME", "receipt-queue"),
		SNSTopicName: getEnv("SNS_TOPIC_NAME", "receipt-topic"),
	}
}

// getEnv returns the environment variable's value if present, or a default
func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
