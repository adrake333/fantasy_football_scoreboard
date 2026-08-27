package main




import (
	"context"
	"flag"
	"net/http"
	"time"
)




func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

//	client := &http.Client{
//		Timeout: 10 * time.Second,
//	}

//	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
//	defer cancel()

//	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://
}
