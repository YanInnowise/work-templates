package main

import (
	"database/sql"
	"fmt"
    "time"
	_ "github.com/lib/pq"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	PostgresUri   string `env:"POSTGRES_URL" envDefault:"postgres://root:pass@localhost:5432/postgres?sslmode=disable"`
	ListenAddress string `env:"LISTEN_ADDRESS" envDefault:":7000"`
	//PostgresHost  string `env:"POSTGRES_HOST" envDefault:":l"`
	//PostgresUser  string `env:"POSTGRES_USER" envDefault:":root"`
	//PostgresPassword string `env:"POSTGRES_PASSWD" envDefault:":qwerty"`
    //PostgresName  string `env:"POSTGRES_NAME" envDefault:":postgres"`

}

var (
	db          *sql.DB
	errorsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gocalc_errors_count",
			Help: "Gocalc Errors Count Per Type",
		},
		[]string{"type"},
	)

	requestsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gocalc_requests_count",
			Help: "Gocalc Requests Count",
		})
)

func main() {
	var err error

	// Initing prometheus
	prometheus.MustRegister(errorsCount)
	prometheus.MustRegister(requestsCount)

	// Getting env
	cfg := config{}
	if err = env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
    
	time.Sleep(time.Second)
	fmt.Println("Sleep over!")
	
	// Connecting to database
	//psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
	//                        cfg.PostgresHost,cfg.ListenAddress,cfg.PostgresUser,cfg.PostgresPassword,cfg.PostgresName)
	
	db, err = sql.Open("postgres",cfg.PostgresUri)
	if err != nil {
		log.Fatalf("Can't connect to postgresql: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Can't ping database: %v", err)
	}

	http.HandleFunc("/", handler)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	requestsCount.Inc()

	keys, ok := r.URL.Query()["q"]
	if !ok || len(keys[0]) < 1 {
		errorsCount.WithLabelValues("missing").Inc()
		log.Println("Url Param 'q' is missing")
		http.Error(w, "Bad Request", 400)
		return
	}
	q := keys[0]
	log.Println("Got query: ", q)

	var result string
	sqlStatement := fmt.Sprintf("SELECT (%s)::numeric", q)
	row := db.QueryRow(sqlStatement)
	err := row.Scan(&result)

	if err != nil {
		log.Println("Error from db: %s", err)
		errorsCount.WithLabelValues("db").Inc()
		http.Error(w, "Internal Server Error", 500)
		return
	}

	fmt.Fprintf(w, "query %s; result %s", q, result)
}
