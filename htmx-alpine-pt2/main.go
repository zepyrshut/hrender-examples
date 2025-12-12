package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/alexedwards/flow"
	"github.com/zepyrshut/hrender"
	"github.com/zepyrshut/hrender/ui"
)

type TodoList struct {
	ID        int
	Name      string
	Completed bool
	Error     string

	ui.CRUDActions
}

var todoList = []TodoList{
	{
		ID:        1,
		Name:      "Sacar a la perra",
		Completed: true,
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

func (t *TodoList) GetID() string {
	return strconv.Itoa(t.ID)
}

func (t *TodoList) SetCRUDActions(actions ui.CRUDActions) {
	t.CRUDActions = actions
}

func main() {
	templatesFS := os.DirFS("./templates")

	h := hrender.NewHTMLRender(templatesFS, false)

	mux := flow.New()

	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("sleeping 0.5 second, path %s, method %s", r.URL.Path, r.Method)
			time.Sleep(200 * time.Millisecond)
			next.ServeHTTP(w, r)
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err := h.RenderW(w, "pages/index", hrender.H{}, "layouts/base")
		if err != nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
		}
	}, "GET")

	mux.HandleFunc("/todo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)

		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})

		err := h.RenderW(w, "fragments/todo-widget", hrender.H{
			"Todo": todoList,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
		}
	}, "GET")

	mux.HandleFunc("/todo", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)

		name := r.FormValue("name")
		if name == "" {
			w.Header().Set("HX-Retarget", "#create-todo-form")
			w.Header().Set("HX-Reswap", "outerHTML")
			err := h.RenderW(w, "fragments/todo-new-form", hrender.H{
				"Name":  name,
				"Error": "El nombre es obligatorio",
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
			}
			return
		}

		newID := 0
		sort.Slice(todoList, func(i, j int) bool {
			return todoList[i].ID > todoList[j].ID
		})
		if len(todoList) > 0 {
			newID = todoList[0].ID + 1
		} else {
			newID = 1
		}

		newItem := TodoList{
			ID:        newID,
			Name:      name,
			Completed: false,
		}

		todoList = append(todoList, newItem)

		err = h.RenderW(w, "fragments/todo-row", hrender.H{
			"ID":        newID,
			"Name":      newItem.Name,
			"Completed": false,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
		}
	}, "POST")

	mux.HandleFunc("/todo/:id/completed", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
			return
		}

		var foundItem *TodoList
		for i := range todoList {
			if todoList[i].ID == id {
				todoList[i].Completed = !todoList[i].Completed
				foundItem = &todoList[i]
				break
			}
		}

		if foundItem == nil {
			http.Error(w, fmt.Sprintf("error loading template, err: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err = h.RenderW(w, "fragments/todo-row", hrender.H{
			"ID":        foundItem.ID,
			"Name":      foundItem.Name,
			"Completed": foundItem.Completed,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}, "PATCH")

	mux.HandleFunc("/todo/:id", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		newName := r.FormValue("name")
		if newName == "" {
			w.Header().Set("HX-Retarget", fmt.Sprintf("#todo-edit-form-%d", id))
			w.Header().Set("HX-Reswap", "outerHTML")
			err := h.RenderW(w, "fragments/todo-edit-form", hrender.H{
				"ID":    id,
				"Name":  newName,
				"Error": "El nombre es obligatorio",
			})
			if err != nil {
				http.Error(w, "error loading template", http.StatusInternalServerError)
			}
			return
		}

		if newName == "some error" {
			log.Println("error triggered")
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var foundItem *TodoList
		for i := range todoList {
			if todoList[i].ID == id {
				todoList[i].Name = newName
				foundItem = &todoList[i]
				break
			}
		}

		if foundItem == nil {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}

		w.Header().Set(hrender.ContentType, hrender.ContentTextHTMLUTF8)
		err = h.RenderW(w, "fragments/todo-row", hrender.H{
			"ID":        foundItem.ID,
			"Name":      foundItem.Name,
			"Completed": foundItem.Completed,
		})
		if err != nil {
			http.Error(w, "error loading template", http.StatusInternalServerError)
		}
	}, "PUT")

	log.Println("server started on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic("server cannot start")
	}
}
