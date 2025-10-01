package handlers

import (
	"fmt"
	"net/http"
)

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
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
}
