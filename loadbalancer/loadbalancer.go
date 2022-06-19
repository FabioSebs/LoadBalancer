package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/TwiN/go-color"
	"github.com/go-co-op/gocron"
)

type Server struct {
	ReverseProxy *httputil.ReverseProxy
	URL          string
	Health       bool
	Name         string
}

type MasterServers struct {
	List   []*Server
	Health bool
	Name   string
}

var (
	index     int32
	portRange = "http://127.0.0.1:808"
)

func (server *Server) testServer() bool {
	resp, err := http.Get(server.URL)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		server.Health = false
		return server.Health
	}

	server.Health = true
	return server.Health
}

func MakeLoadBalaner(servers int) {
	master := MasterServers{}
	index = 0
	for i := 0; i < servers; i++ {
		master.List = append(master.List, createServer())
	}
	index = 0
	http.HandleFunc("/", makeRequest(&master))
	healthCheck(&master)
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func makeRequest(servers *MasterServers) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		i := int(index) % len(servers.List)
		revproxy := servers.List[i]
		index++
		revproxy.ReverseProxy.ServeHTTP(w, r)
	}
}

func createServer() *Server {
	link := portRange + strconv.Itoa(int(index))
	endpoint, _ := url.Parse(link)
	revproxy := httputil.NewSingleHostReverseProxy(endpoint)
	server := Server{
		ReverseProxy: revproxy,
		URL:          link,
		Health:       true,
		Name:         fmt.Sprintf("Server %d", index+1),
	}
	index++
	return &server
}

func healthCheck(master *MasterServers) {
	scheduler := gocron.NewScheduler(time.Local)
	for _, server := range master.List {
		_, err := scheduler.Every(5).Second().Do(func(s *Server) {
			if s.testServer() {
				log.Printf(color.Colorize(color.Green, "%s is running healthy\n"), s.Name)
			} else {
				log.Printf(color.Colorize(color.Red, "%s is NOT healthy\n"), s.Name)
			}
		}, server)
		if err != nil {
			log.Fatal(err)
		}
	}
	scheduler.StartAsync()
}
