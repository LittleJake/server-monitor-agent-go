package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	VERSION               = "Alpha-20250311.1-golang"
	LOG_LEVEL             string
	HOST                  string
	PORT                  string
	SSL                   bool
	REPORT_ONCE           bool
	SOCKET_TIMEOUT        int
	SERVER_URL            string
	SERVER_TOKEN          string
	REPORT_MODE           string
	UUID                  string
	USER_AGENT            = VERSION + " +https://github.com/LittleJake/server-monitor-agent-go"
	SERVER_URL_INFO       string
	SERVER_URL_COLLECTION string
	SERVER_URL_HASH       string
	SERVER_URL_COMMAND    string
	IPV4                  string
	IPV6                  string
	IPV4_API              string
	IPV6_API              string
	COUNTRY               map[string]string
)

func loadUUID() string {
	// Load UUID from file
	file, err := os.ReadFile(".uuid")
	if err != nil {
		//generate new UUID
		newUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
		err := os.WriteFile(".uuid", []byte(newUUID), 0644)
		if err != nil {
			log.Fatalf("Error writing UUID: %v", err)
		}
		return newUUID
	}
	return strings.TrimSpace(string(file))
}

func init() {

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	HOST = getEnv("HOST", "localhost")
	PORT = getEnv("PORT", "6379")
	SSL, _ = strconv.ParseBool(getEnv("SSL", "false"))
	REPORT_ONCE, _ = strconv.ParseBool(getEnv("REPORT_ONCE", "false"))
	SOCKET_TIMEOUT, _ = strconv.Atoi(getEnv("SOCKET_TIMEOUT", "10"))
	SERVER_URL = getEnv("SERVER_URL", "http://localhost:8000")
	REPORT_MODE = strings.ToLower(getEnv("REPORT_MODE", "redis"))
	SERVER_TOKEN = getEnv("SERVER_TOKEN", "")
	LOG_LEVEL = getEnv("LOG_LEVEL", "INFO")
	UUID = loadUUID()

	SERVER_URL_INFO = fmt.Sprintf("%s/api/report/info/%s", SERVER_URL, UUID)
	SERVER_URL_COLLECTION = fmt.Sprintf("%s/api/report/collection/%s", SERVER_URL, UUID)
	SERVER_URL_HASH = fmt.Sprintf("%s/api/report/hash/%s", SERVER_URL, UUID)
	SERVER_URL_COMMAND = fmt.Sprintf("%s/api/report/command/%s", SERVER_URL, UUID)

	IPV4_API = getEnv("IPV4_API", "https://api.ipify.org")
	IPV6_API = getEnv("IPV6_API", "https://api6.ipify.org")

	// Initialize logger
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	setLogLevel(LOG_LEVEL)

	getIP()
	getCountry()
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getRedisConn() redis.Conn {
	// Connect to Redis
	conn, _ := redis.Dial("tcp", fmt.Sprintf("%v:%v", HOST, PORT), redis.DialUseTLS(SSL))

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

func getInfo() map[string]interface{} {

	info := map[string]interface{}{
		"Connection":     getConnections(),
		"Country":        COUNTRY["country_name"],
		"Country Code":   COUNTRY["country_code"],
		"CPU":            getCPUInfo(),
		"IPV4":           replaceString(IPV4, "\\d*\\.\\d*\\.\\d*", "*.*.*"),
		"IPV6":           replaceString(IPV6, "[a-fA-F0-9]*:", "*:"),
		"Load Average":   getLoadAvg(),
		"Process":        getProcessNum(),
		"System Version": getSysVersion(),
		"Throughput":     getThroughput(),
		"Update Time":    time.Now().Unix(),
		"Uptime":         getUptime(),
		"Agent Version":  VERSION,
	}

	return info
}

func postRequest(url string, headers map[string]string, data string) (string, error) {

	// Post data to the server
	client := &http.Client{
		Timeout: time.Duration(SOCKET_TIMEOUT) * time.Second,
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Error creating request: %v", err))
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Error post data: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Error reading response body: %v", err))
		return "", err
	}

	logMessage(DEBUG, string(body))
	return string(body), nil
}

func getRequest(url string, headers map[string]string) (string, error) {
	// Post data to the server
	client := &http.Client{
		Timeout: time.Duration(SOCKET_TIMEOUT) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Error creating request: %v", err))
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Fail to get data: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logMessage(ERROR, fmt.Sprintf("Error reading response body: %v", err))
		return "", err
	}

	logMessage("DEBUG", string(body))
	return string(body), nil
}

func replaceString(input, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(input, replacement)
}

func getIP() {
	var err error
	IPV4, err = getRequest(IPV4_API, map[string]string{})
	if err != nil {
		IPV4 = "None"
	}
	IPV6, err = getRequest(IPV6_API, map[string]string{})
	if err != nil {
		IPV6 = "None"
	}

	logMessage(INFO, IPV4)
	logMessage(INFO, IPV6)
}

func getCountry() {
	COUNTRY = map[string]string{
		"country_name": "Unknown",
		"country_code": "Unknown",
	}

	data, err := getRequest("https://ip-api.io/json", map[string]string{})
	if err != nil {
		logMessage(ERROR, "Fail to get country")
		return
	}

	logMessage(DEBUG, data)

	var country map[string]interface{}
	err = json.Unmarshal([]byte(data), &country)
	if err != nil {
		logMessage(ERROR, "Fail to get country")
		return
	}
	COUNTRY = map[string]string{
		"country_name": country["country_name"].(string),
		"country_code": country["country_code"].(string),
	}

	if strings.Contains(country["country_name"].(string), "Hong Kong") ||
		strings.Contains(country["country_name"].(string), "Hong Kong") {
		COUNTRY["country_name"] = country["country_name"].(string) + ", SAR"
	} else if strings.Contains(country["country_name"].(string), "Taiwan") {
		COUNTRY["country_name"] = country["country_name"].(string) + " Province"
		COUNTRY["country_code"] = "CN"
	}

}

func report() {
	logMessage(INFO, "Start Reporting")
	aggregateStat := getAggregateStat()
	info := getInfo()

	jsonInfo, err := json.Marshal(info)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Fail to get data: %v", err))
		return
	}
	jsonAggregateStat, _ := json.Marshal(aggregateStat)

	logMessage(DEBUG, string(jsonAggregateStat))
	logMessage(DEBUG, string(jsonInfo))

	if REPORT_MODE == "redis" {
	}
	if REPORT_MODE == "http" {
		if SERVER_TOKEN == "" {
			log.Fatalf("Please generate server token using `php think token add --uuid %s`", UUID)
		}
		postRequest(SERVER_URL_HASH, map[string]string{"User-Agent": USER_AGENT, "Content-Type": "application/json", "authorization": SERVER_TOKEN}, "{\"ip\": \"none\"}")
		postRequest(SERVER_URL_INFO, map[string]string{"User-Agent": USER_AGENT, "Content-Type": "application/json", "authorization": SERVER_TOKEN}, string(jsonInfo))
		postRequest(SERVER_URL_COLLECTION, map[string]string{"User-Agent": USER_AGENT, "Content-Type": "application/json", "authorization": SERVER_TOKEN}, string(jsonAggregateStat))

		logMessage("INFO", "Finish Reporting")
	}
}

func main() {
	for {
		report()
		// getInfo()
		if !REPORT_ONCE {
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}
}
