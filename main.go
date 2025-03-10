package main

// func init() {
// 	// Load environment variables
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}

// }

func main() {
	get_cpu()
	get_uptime()
	get_connections()
	get_sys_version()
	get_load()
	get_load_avg()
	get_mem_info()
}
