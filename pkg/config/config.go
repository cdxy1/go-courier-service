package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type Сonfig struct {
	Port             string
	Postgres         *PostgresConfig
	OrderServiceGRPC string
	OrderServiceHTTP string
	Kafka            *KafkaConfig
	OrderPolling     bool
	Delivery         *DeliveryConfig
	Pprof            *PprofConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Db       string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
	Version string
	Enabled bool
}

type DeliveryConfig struct {
	MonitorInterval time.Duration
	OnFootDuration  time.Duration
	ScooterDuration time.Duration
	CarDuration     time.Duration
}

type PprofConfig struct {
	Enabled       bool
	Host          string
	Port          string
	BasicUser     string
	BasicPassword string
}

func (p *PostgresConfig) GetURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", p.User, p.Password, p.Host, p.Port, p.Db)
}

func GetEnv() *Сonfig {
	if err := godotenv.Load(".env", ".env.example"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("failed to load .env file: %v\n", err)
	}

	port := getServicePort()
	postgres := getPostgresConfig()
	orderServiceGRPC := getOrderServiceGRPC()
	orderServiceHTTP := getOrderServiceHTTP()
	kafka := getKafkaConfig()
	orderPolling := getOrderPolling()
	delivery := getDeliveryConfig()
	pprof := getPprofConfig()

	return &Сonfig{
		Port:             port,
		Postgres:         postgres,
		OrderServiceGRPC: orderServiceGRPC,
		OrderServiceHTTP: orderServiceHTTP,
		Kafka:            kafka,
		OrderPolling:     orderPolling,
		Delivery:         delivery,
		Pprof:            pprof,
	}
}

func getServicePort() string {
	var port string

	inputPort := pflag.String("port", "", "server port")
	pflag.Parse()

	port = *inputPort
	if port == "" {
		port = os.Getenv("PORT")
	}

	return port
}

func getPostgresConfig() *PostgresConfig {
	return &PostgresConfig{Host: os.Getenv("POSTGRES_HOST"), Port: os.Getenv("POSTGRES_PORT"), User: os.Getenv("POSTGRES_USER"), Password: os.Getenv("POSTGRES_PASSWORD"), Db: os.Getenv("POSTGRES_DB")}
}

func getOrderServiceGRPC() string {
	addr := os.Getenv("ORDER_SERVICE_HOST")
	if addr == "" {
		return "localhost:50051"
	}
	return addr
}

func getOrderServiceHTTP() string {
	addr := os.Getenv("ORDER_SERVICE_HTTP")
	if addr == "" {
		return "http://localhost:8083"
	}
	return addr
}

func getKafkaConfig() *KafkaConfig {
	rawBrokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	brokers := splitCSV(rawBrokers)
	enabled := len(brokers) > 0
	if v := strings.TrimSpace(os.Getenv("KAFKA_ENABLED")); v != "" {
		v = strings.ToLower(v)
		enabled = v == "true" || v == "1" || v == "yes"
	}

	return &KafkaConfig{
		Brokers: brokers,
		Topic:   strings.TrimSpace(os.Getenv("KAFKA_ORDER_TOPIC")),
		GroupID: strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_GROUP")),
		Version: strings.TrimSpace(os.Getenv("KAFKA_VERSION")),
		Enabled: enabled,
	}
}

func getOrderPolling() bool {
	value := strings.TrimSpace(os.Getenv("ORDER_POLLING_ENABLED"))
	if value == "" {
		return false
	}
	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes"
}

func getDeliveryConfig() *DeliveryConfig {
	return &DeliveryConfig{
		MonitorInterval: getDuration("DELIVERY_MONITOR_INTERVAL", time.Second*10),
		OnFootDuration:  getDuration("DELIVERY_DURATION_ON_FOOT", time.Minute*30),
		ScooterDuration: getDuration("DELIVERY_DURATION_SCOOTER", time.Minute*15),
		CarDuration:     getDuration("DELIVERY_DURATION_CAR", time.Minute*5),
	}
}

func getPprofConfig() *PprofConfig {
	enabled := strings.TrimSpace(os.Getenv("PPROF_ENABLED"))
	pprofEnabled := false
	if enabled != "" {
		enabled = strings.ToLower(enabled)
		pprofEnabled = enabled == "true" || enabled == "1" || enabled == "yes"
	}

	host := strings.TrimSpace(os.Getenv("PPROF_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	port := strings.TrimSpace(os.Getenv("PPROF_PORT"))
	if port == "" {
		port = "6060"
	}

	return &PprofConfig{
		Enabled:       pprofEnabled,
		Host:          host,
		Port:          port,
		BasicUser:     strings.TrimSpace(os.Getenv("PPROF_BASIC_USER")),
		BasicPassword: strings.TrimSpace(os.Getenv("PPROF_BASIC_PASSWORD")),
	}
}

func getDuration(envName string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
