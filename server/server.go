package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/FabioSebs/LoadBalancer/content"
	"github.com/gorilla/mux"
)

type ServerSlice struct {
	Ports []int
}

type Template struct {
	Number int
}

func (s *ServerSlice) Populate(number int) {
	if number >= 10 {
		fmt.Println("Number of Ports can't exceed 10")
		return
	}

	for x := 0; x < number; x++ {
		s.Ports = append(s.Ports, x)
	}
}

func (s *ServerSlice) Pop() int {
	el := s.Ports[0]
	s.Ports = s.Ports[1:]
	return el
}

func RunManyServers(servers int) {
	ss := ServerSlice{}
	ss.Populate(servers)
	var wg sync.WaitGroup

	wg.Add(servers)
	for i := 0; i < servers; i++ {
		go makeServer(&ss, wg)
	}

	wg.Wait()
}

func makeServer(ss *ServerSlice, wg sync.WaitGroup) {
	port := ss.Pop()
	defer wg.Done()
	r := mux.NewRouter()
	s := http.Server{Addr: fmt.Sprintf(":808%d", port), Handler: r}
	r.HandleFunc("/", returnSomething(port)) 
	r.HandleFunc("/shutdown", shutDown(&s))
	fmt.Println("Running on server 808" + strconv.Itoa(port))

	s.ListenAndServe()
}

func returnSomething(serverNo int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content Type", "text/html")
		doc := content.HTML
		templates := template.New("template")
		templates.New("doc").Parse(doc)

		context := Template{
			Number: serverNo + 1,
		}
		templates.Lookup("doc").Execute(w, context)
	}
}

func shutDown(server *http.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("400 - Server Shut Down!"))
		server.Shutdown(context.Background())
	}
}
