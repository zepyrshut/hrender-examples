package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/zepyrshut/hrender"
)

type TodoList struct {
	ID        int
	Name      string
	Completed bool
}

var todoList = []TodoList{
	{
		ID:        1,
		Name:      "Sacar a la perra",
		Completed: false,
	},
	{
		ID:        2,
		Name:      "Limpiar la casa",
		Completed: false,
	},
	{
		ID:        3,
		Name:      "Ir al supermercado",
		Completed: false,
	},
}

func main() {
	templatesFS := os.DirFS("./templates")

	h := hrender.NewHTMLRender(templatesFS, false)

	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err := h.Render(w, "pages/index", hrender.H{})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	mux.Handle("GET /todo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)

		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})

		err := h.Render(w, "fragments/todo-widget", hrender.H{
			"TodoList": todoList,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	mux.Handle("GET /todo/new", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err := h.Render(w, "fragments/todo-row-edit", hrender.H{})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	mux.Handle("POST /todo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		newID := 0
		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})
		newID = todoList[0].ID + 1

		log.Println(todoList[len(todoList)-1].ID)
		log.Println(todoList)

		newItem := TodoList{
			ID:        newID,
			Name:      r.FormValue("name"),
			Completed: false,
		}

		todoList = append(todoList, newItem)

		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err = h.Render(w, "fragments/todo-row", hrender.H{
			"ID":        newID,
			"Name":      newItem.Name,
			"Completed": false,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	mux.Handle("PATCH /todo/{id}/completed", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var foundItem *TodoList
		for i := range todoList {
			if todoList[i].ID == id {
				todoList[i].Completed = true
				foundItem = &todoList[i]
				break
			}
		}

		if foundItem == nil {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}

		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err = h.Render(w, "fragments/todo-row", hrender.H{
			"ID":        foundItem.ID,
			"Name":      foundItem.Name,
			"Completed": true,
		})
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
