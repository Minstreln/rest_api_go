package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// type user struct {
// 	Name string `json:"name"`
// 	Age  string `json:"age"`
// 	City string `json:"city"`
// }

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello root route")
	w.Write([]byte("Hello Root route"))
	fmt.Println("Hello Root route")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	// teachers/{9}
	// teachers/9
	switch r.Method {
	case http.MethodGet:
		fmt.Println(r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.TrimSuffix(path, "/")

		fmt.Println("The ID is:", userID)

		fmt.Println("Query Params", r.URL.Query())
		queryParams := r.URL.Query()
		sortBy := queryParams.Get("sortBy")
		key := queryParams.Get("key")
		sortorder := queryParams.Get("sortorder")

		if sortorder == "" {
			sortorder = "DESC"
		}

		fmt.Printf("sortBy: %v, key: %v, sortorder: %v", sortBy, key, sortorder)

		w.Write([]byte("Hello GET Method on Teachers route"))
		// fmt.Println("Hello GET Method on Teachers route")
		return
	case http.MethodPost:
		w.Write([]byte("Hello POST Method on Teachers route"))
		fmt.Println("Hello POST Method on Teachers route")
		return
	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Teachers route"))
		fmt.Println("Hello PUT Method on Teachers route")
		return
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Teachers route"))
		fmt.Println("Hello PATCH Method on Teachers route")
		return
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Teachers route"))
		fmt.Println("Hello DELETE Method on Teachers route")
		return
	}
	// w.Write([]byte("Hello Teachers route"))
	// fmt.Println("Hello Teachers route")
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET Method on Students route"))
		fmt.Println("Hello GET Method on Students route")
		return
	case http.MethodPost:
		w.Write([]byte("Hello POST Method on Students route"))
		fmt.Println("Hello POST Method on Students route")
		return
	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Students route"))
		fmt.Println("Hello PUT Method on Students route")
		return
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Students route"))
		fmt.Println("Hello PATCH Method on Students route")
		return
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Students route"))
		fmt.Println("Hello DELETE Method on Students route")
		return
	}
	w.Write([]byte("Hello Students route"))
	fmt.Println("Hello Students route")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET Method on Execs route"))
		fmt.Println("Hello GET Method on Execs route")
		return
	case http.MethodPost:
		w.Write([]byte("Hello POST Method on Execs route"))
		fmt.Println("Hello POST Method on Execs route")
		return
	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Execs route"))
		fmt.Println("Hello PUT Method on Execs route")
		return
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Execs route"))
		fmt.Println("Hello PATCH Method on Execs route")
		return
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Execs route"))
		fmt.Println("Hello DELETE Method on Execs route")
		return
	}
	w.Write([]byte("Hello Execs route"))
	fmt.Println("Hello Execs route")
}

func main() {
	port := ":3000"

	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teachersHandler)

	http.HandleFunc("/students/", studentsHandler)

	http.HandleFunc("/execs/", execsHandler)

	fmt.Println("Server is running on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
