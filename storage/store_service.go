package store

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

// Define the struct wrapper around raw Redis client
type StorageService struct {
	redisClient    *redis.Client
	PostgresClient *gorm.DB
}
type urls struct {
	gorm.Model
	Tinyurl string `gorm:"unique;not null"`
	Longurl string
}
type Config struct {
	REDIS_HOST        string `mapstructure:"REDIS_HOST"`
	REDIS_PASSWORD    string `mapstructure:"REDIS_PASSWORD"`
	REDIS_DB          string `mapstructure:"REDIS_DB"`
	POSTGRES_HOST     string `mapstructure:"POSTGRES_HOST"`
	POSTGRES_PORT     string `mapstructure:"POSTGRES_PORT"`
	POSTGRES_USER     string `mapstructure:"POSTGRES_USER"`
	POSTGRES_DB       string `mapstructure:"POSTGRES_DB"`
	POSTGRES_PASSWORD string `mapstructure:"POSTGRES_PASSWORD"`
	POSTGRES_SSLMODE  string `mapstructure:"POSTGRES_SSLMODE"`
}

// Top level declarations for the storeService and Redis context
var (
	storeService = &StorageService{}
	ctx          = context.Background()
)

// Note that in a real world usage, the cache duration shouldn't have
// an expiration time, an LRU policy config should be set where the
// values that are retrieved less often are purged automatically from
// the cache and stored back in RDBMS whenever the cache is full

const CacheDuration = 6 * time.Hour

// Initializing the store service and return a store pointer
func InitializeStore() *StorageService {

	config, err := LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	db, _ := strconv.Atoi(config.REDIS_DB)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST,
		Password: config.REDIS_PASSWORD,
		DB:       db,
	})

	pong, err := redisClient.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	connection_string := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", config.POSTGRES_HOST, config.POSTGRES_PORT, config.POSTGRES_USER, config.POSTGRES_DB, config.POSTGRES_PASSWORD, config.POSTGRES_SSLMODE)
	dbClient, err := gorm.Open("postgres", connection_string)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\nSuccessfully connected to PGSQL\n\n")
	dbClient.AutoMigrate(&urls{})
	storeService.PostgresClient = dbClient
	return storeService
}

func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	storeService.PostgresClient.Create(&urls{Tinyurl: shortUrl, Longurl: originalUrl})
	err := storeService.redisClient.Set(shortUrl, originalUrl, CacheDuration).Err()
	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}

}

func RetrieveInitialUrl(shortUrl string) string {
	result, err := storeService.redisClient.Get(shortUrl).Result()
	if err != nil {
		var url urls
		storeService.PostgresClient.Where("tinyurl = ?", shortUrl).Select("longurl").Find(&url)
		fmt.Println("Not in redis, searching in pgsql.......")
		if url.Longurl != "" {
			err := storeService.redisClient.Set(url.Tinyurl, url.Longurl, CacheDuration).Err()
			if err != nil {
				panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, url.Tinyurl, url.Longurl))
			}
			return url.Longurl
		} else {
			panic("Short url not recognized.")
		}
	}

	return result
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("prod")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
