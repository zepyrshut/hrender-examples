package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/starfederation/datastar-go/datastar"
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

type TodoSignals struct {
	NewTodoName string `json:"newtodo"`
}

func main() {
	templatesFS := os.DirFS("./templates")

	h := hrender.NewHTMLRender(templatesFS, false)

	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err := h.RenderW(w, "pages/index", hrender.H{})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}))

	mux.Handle("GET /todo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})

		templ, err := h.RenderS("fragments/todo-widget", hrender.H{
			"TodoList": todoList,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}

		sse := datastar.NewSSE(w, r)
		if err := sse.PatchElements(templ); err != nil {
			log.Println("error", err)
		}
	}))

	mux.Handle("GET /todo/new", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ, err := h.RenderS("fragments/todo-row-edit", hrender.H{})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}

		sse := datastar.NewSSE(w, r)
		if err := sse.PatchElements(templ,
			datastar.WithModePrepend(),
			datastar.WithSelectorID("todo-list-body"),
		); err != nil {
			log.Println("error", err)
		}
	}))

	mux.Handle("POST /todo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signals TodoSignals
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		newID := 0
		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})
		newID = todoList[0].ID + 1

		newItem := TodoList{
			ID:        newID,
			Name:      signals.NewTodoName,
			Completed: false,
		}

		todoList = append(todoList, newItem)

		templ, _ := h.RenderS("fragments/todo-row", hrender.H{
			"ID":        newID,
			"Name":      newItem.Name,
			"Completed": false,
		})

		sse := datastar.NewSSE(w, r)

		if err := sse.PatchElements(templ,
			datastar.WithModeReplace(),
			datastar.WithSelectorID("new-todo-row"),
		); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		signals.NewTodoName = ""
		sse.MarshalAndPatchSignals(signals)
	}))

	mux.Handle("PATCH /todo/{id}/completed", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("patch todo triggered")

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

		templ, err := h.RenderS("fragments/todo-row", hrender.H{
			"ID":        foundItem.ID,
			"Name":      foundItem.Name,
			"Completed": true,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}

		sse := datastar.NewSSE(w, r)
		if err := sse.PatchElements(templ, datastar.WithModeOuter(),
			datastar.WithSelectorID(fmt.Sprintf("todo-%d", id)),
		); err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
		}
	}))

	log.Println("server started on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(fmt.Sprintf("server cannot start %v", err))
	}
}
