package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type todos struct {
	Id          int    `json:"id"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Completed   bool   `json:"Completed"`
	CreatedAt   string `json:"CreatedAt"`
}

var db *sql.DB

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome! :D\n"))
}

func getAllTodos(w http.ResponseWriter, r *http.Request) {
	// SQL Query
	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		checkError(err)
		return
	}
	defer rows.Close()

	var todosArray []todos
	for rows.Next() {
		var newtodos todos
		err = rows.Scan(&newtodos.Id, &newtodos.Title, &newtodos.Description, &newtodos.Completed, &newtodos.CreatedAt)
		if err != nil {
			checkError(err)
			return
		}
		todosArray = append(todosArray, newtodos)
	}

	jsonData, err := json.Marshal(todosArray)
	if err != nil {
		checkError(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func getTodosID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// SQL query
	var newtodos todos
	row := db.QueryRow("SELECT * FROM todos where id=?", id)
	if err := row.Scan(&newtodos.Id, &newtodos.Title, &newtodos.Description, &newtodos.Completed, &newtodos.CreatedAt); err != nil {
		checkError(err)
		return
	}

	jsonData, err := json.Marshal(newtodos)
	if err != nil {
		checkError(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	// Get POST data?
	var newtodos todos

	err := json.NewDecoder(r.Body).Decode(&newtodos)
	if err != nil {
		checkError(err)
		return
	}

	sql := `INSERT INTO todos(Title, Description, Completed) VALUES (?, ?, ?)`
	result, err := db.Exec(sql, newtodos.Title, newtodos.Description, newtodos.Completed)
	if err != nil {
		checkError(err)
		return
	}
	// sql.Result's LastInsertId() obtain AUTO_INCREMENT values
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		checkError(err)
		return
	}

	ret := fmt.Sprintf("You create a new todo with %d ID", lastInsertID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ret + "\n"))
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	// Seem like must use JSON
	var newtodos todos

	params := mux.Vars(r)
	id := params["id"]
	err := json.NewDecoder(r.Body).Decode(&newtodos)
	if err != nil {
		checkError(err)
		return
	}
	// The newtodos will contains the new data
	sql := "UPDATE todos SET Title=?, Description=?, Completed=?, CreatedAt=? where id=?"
	_, err = db.Exec(sql, newtodos.Title, newtodos.Description, newtodos.Completed, newtodos.CreatedAt, id)
	if err != nil {
		checkError(err)
		return
	}
	respondWithJSON(w, r, http.StatusOK, newtodos)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	// Get delete ID
	params := mux.Vars(r)
	id := params["id"]

	// Delete
	_, err := db.Exec("DELETE FROM todos where id=?", id)
	if err != nil {
		checkError(err)
		return
	}

	ret := fmt.Sprintf("Successfully delete %s ID todos", id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ret + "\n"))
}

// Database Data
const (
	UserName string = "root"
	Password string = ""
	Addr     string = "127.0.0.1"
	Port     int    = 3306
	Database string = "test"
)

func CreateTable(db *sql.DB) {
	sql := `CREATE TABLE IF NOT EXISTS todos(
		ID bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
		Title VARCHAR(255),
		Description VARCHAR(512),
		Completed INT(2) NOT NULL DEFAULT 0,
		CreatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(sql); err != nil {
		checkError(err)
		return
	}
	fmt.Println("Create Table Success!")
}

func main() {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", UserName, Password, Addr, Port, Database)
	db, _ = sql.Open("mysql", conn)

	CreateTable(db)

	r := mux.NewRouter()
	r.HandleFunc("/", Index).Methods("GET")
	r.HandleFunc("/todos", getAllTodos).Methods("GET")
	r.HandleFunc("/todos/{id}", getTodosID).Methods("GET")
	r.HandleFunc("/todos", createTodo).Methods("POST")
	r.HandleFunc("/todos/{id}", updateTodo).Methods("PUT")
	r.HandleFunc("/todos/{id}", deleteTodo).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":9090", r))

	fmt.Println("bye!")
	db.Close()
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	result, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(result)
}

func checkError(err error) {
	fmt.Printf("Found exception: %v\n", err)
}
