package http

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/matryer/resync"
)

var once resync.Once
var pool *HttpClientPool

type HttpClientPool struct {
	client *http.Client
}

func GetHttpClientPool() *HttpClientPool {
	if pool == nil {
		once.Do(create)
	}
	return pool
}

func create() {
	var proxy func(*http.Request) (*url.URL, error) = nil
	if proxyEnv := os.Getenv("HTTPS_PROXY"); proxyEnv != "" {
		proxyURL, _ := url.Parse(proxyEnv)
		proxy = http.ProxyURL(proxyURL)
	}

	pool = &HttpClientPool{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:               proxy,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				// DisableKeepAlives:   true,
				// TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				// DialContext: (&net.Dialer{
				// 	Timeout:   5 * time.Second,
				// 	DualStack: false, // 禁用 IPv6
				// }).DialContext,
			},
			Timeout: 30 * time.Second,
		},
	}
}

func (p *HttpClientPool) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "criptobox/0.1")

	return req, nil
}

func (p *HttpClientPool) Do(req *http.Request) ([]byte, error) {
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// func main() {
// 	pool := NewHttpClientPool()
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	// 优雅退出处理
// 	go func() {
// 		sigChan := make(chan os.Signal, 1)
// 		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 		<-sigChan
// 		cancel()
// 	}()

// 	var wg sync.WaitGroup
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			for ctx.Err() == nil {
// 				data, err := pool.Get("https://example.com")
// 				if err != nil {
// 					log.Printf("Worker %d error: %v", id, err)
// 					continue
// 				}
// 				log.Printf("Worker %d got %d bytes", id, len(data))
// 				time.Sleep(1 * time.Second)
// 			}
// 		}(i)
// 	}
// 	wg.Wait()
// }
