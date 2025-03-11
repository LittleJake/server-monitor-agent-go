package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

var VERSION = "Alpha-20250311.1-golang"
var HOST = getEnv("HOST", "localhost")
var PORT = getEnv("PORT", "6379")
var SSL = getEnv("SSL", "false")

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return strings.ToLower(defaultValue)
	}
	return strings.ToLower(value)
}

func getRedisConn() redis.Conn {
	// Connect to Redis
	conn, _ := redis.Dial("tcp", fmt.Sprintf("%v:%v", HOST, PORT), redis.DialUseTLS(SSL == "true"))

	return conn
}

func getAggregateStat() map[string]interface{} {
	aggregateStat := map[string]interface{}{
		"Battery": json.RawMessage("{}"),
		"Disk":    json.RawMessage(getDiskInfo()),
		"Fan":     json.RawMessage("{}"),
		"IO":      json.RawMessage(getIO()),
		"Load":    json.RawMessage(getLoad()),
		"Memory":  json.RawMessage(getMemInfo()),
		"Network": json.RawMessage(getNetwork()),
		"Ping":    json.RawMessage("{}"),
		"Thermal": json.RawMessage("{}"),
	}

	return aggregateStat
}

func main() {
	info := map[string]interface{}{
		"Connection":     getConnections(),
		"Country":        "",
		"Country Code":   "",
		"CPU":            getCPUInfo(),
		"IPV4":           "None",
		"IPV6":           "None",
		"Load Average":   getLoadAvg(),
		"Process":        getProcessNum(),
		"System Version": getSysVersion(),
		"Throughput":     getThroughput(),
		"Update Time":    time.Now().Format("2006-01-02 15:04:05"),
		"Uptime":         getUptime(),
		"Agent Version":  VERSION,
	}

	getAggregateStat()

	time.Sleep(5 * time.Second)

	aggregateStat := getAggregateStat()

	jsonInfo, err := json.Marshal(info)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	jsonAggregateStat, _ := json.Marshal(aggregateStat)

	log.Printf("Info: %s\n", jsonInfo)
	log.Printf("Aggregate Stat: %s\n", jsonAggregateStat)

}
