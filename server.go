package cddamap

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func CreateRouter(server *HTTPServer) (*mux.Router, error) {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/api/worlds": server.GetWorlds,
		},
		"POST": {},
		"PUT":  {},
		"OPTIONS": {
			"": options,
		},
	}

	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHttpHandler(localMethod, localRoute, localHandler)

			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
}

func makeHttpHandler(localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeCorsHeaders(w, r)
		if err := handlerFunc(w, r, mux.Vars(r)); err != nil {
			httpError(w, err)
		}
	}
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

type HTTPServer struct {
	DB *DB
}

func NewHTTPServer(db *DB) *HTTPServer {
	s := &HTTPServer{
		DB: db,
	}

	return s
}

func writeJSON(w http.ResponseWriter, code int, thing interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	val, err := json.Marshal(thing)
	w.Write(val)
	return err
}

func writeJSONDirect(w http.ResponseWriter, code int, thing []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(thing)
}

func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	if err != nil {
		log.WithField("err", err).Error("http error")
		http.Error(w, err.Error(), statusCode)
	}
}

func options(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *HTTPServer) GetWorlds(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	worlds, err := s.DB.GetWorlds()

	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, worlds)

	return nil
}
