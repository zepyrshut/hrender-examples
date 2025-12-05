package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zepyrshut/hrender"
)

func main() {
	templatesFS := os.DirFS("./templates")

	h := hrender.NewHTMLRender(templatesFS, false)

	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err := h.Render(w, "index", hrender.H{})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	log.Println("server started on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic("server cannot start")
	}
}
