package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cfgListen     = ":8099"
	cfgUrl        = ""
	cfgCfAPIToken = ""
)

// custom collector for updating values without direct processing, such as loops
type MyCollector struct {
	minutesDesc      *prometheus.Desc
	minutesTotalDesc *prometheus.Desc
}

// разметка типов для парсинга json
type Response struct {
	Result struct {
		Creator                  string  `json:"creator"`
		VideoCount               int     `json:"videoCount"`
		TotalStorageMinutes      float64 `json:"totalStorageMinutes"`
		TotalStorageMinutesLimit float64 `json:"totalStorageMinutesLimit"`
	} `json:"result"`
	Success  bool  `json:"success"`
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
}

// Describe implements MyCollector
func (c *MyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.minutesDesc
	ch <- c.minutesTotalDesc
}

// Collect implements MyCollector
func (c *MyCollector) Collect(ch chan<- prometheus.Metric) {
	// Create a new request using http and handle error
	req, err := http.NewRequest("GET", cfgUrl, nil)
	if err != nil {
		fmt.Println("No response from request")
	}

	// add authorization header to the req
	req.Header.Add("Authorization", "bearer "+cfgCfAPIToken)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	//log.Println(string([]byte(body)))

	var query Response
	if err := json.Unmarshal(body, &query); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	value := query.Result.TotalStorageMinutes // Your code to fetch the counter value goes here.
	ch <- prometheus.MustNewConstMetric(
		c.minutesDesc,
		prometheus.GaugeValue,
		value,
	)
	value1 := query.Result.TotalStorageMinutesLimit
	ch <- prometheus.MustNewConstMetric(
		c.minutesTotalDesc,
		prometheus.CounterValue,
		value1,
	)
}

func NewMyCollector() *MyCollector {
	return &MyCollector{
		minutesDesc:      prometheus.NewDesc("stream_capacity", "Занято в минутах", nil, nil),
		minutesTotalDesc: prometheus.NewDesc("stream_limit", "Текущий лимит", nil, nil),
	}
}

func main() {

	flag.StringVar(&cfgListen, "api_port", cfgListen, "exporter port")
	flag.StringVar(&cfgUrl, "api_url", os.Getenv("API_URL"), "cloudflare url for api")
	flag.StringVar(&cfgCfAPIToken, "api_key", os.Getenv("API_KEY"), "cloudflare api key")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(NewMyCollector())

	Mux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	Mux.Handle("/metrics", promHandler)

	func() {
		log.Fatal(http.ListenAndServe(cfgListen, Mux))
	}()
}
