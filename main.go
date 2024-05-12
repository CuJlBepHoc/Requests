package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type User struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Friends []int64 `json:"friends"`
}

var users = make(map[int64]User)
var nextUserID int64 = 1

func main() {
	router := chi.NewRouter()
	router.Post("/create", createUser)
	router.Post("/make_friends", makeFriends)
	router.Delete("/user/{id}", deleteUser)
	router.Get("/friends/{id}", getFriends)
	router.Put("/user/{id}", updateAge)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.ID = nextUserID
	users[nextUserID] = user
	nextUserID++

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": user.ID})
}

func makeFriends(w http.ResponseWriter, r *http.Request) {
	var data struct {
		SourceID int64 `json:"source_id"`
		TargetID int64 `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceUser, ok := users[data.SourceID]
	if !ok {
		http.Error(w, "Source user not found", http.StatusNotFound)
		return
	}

	targetUser, ok := users[data.TargetID]
	if !ok {
		http.Error(w, "Target user not found", http.StatusNotFound)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, targetUser.ID)
	targetUser.Friends = append(targetUser.Friends, sourceUser.ID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s and %s are now friends\n", sourceUser.Name, targetUser.Name)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	for _, friendID := range user.Friends {
		friend, _ := users[friendID]
		for i, id := range friend.Friends {
			if id == user.ID {
				friend.Friends = append(friend.Friends[:i], friend.Friends[i+1:]...)
				break
			}
		}
		users[friend.ID] = friend
	}

	delete(users, userID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s has been deleted\n", user.Name)
}

func getFriends(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	friends := make([]string, len(user.Friends))
	for i, friendID := range user.Friends {
		friend, _ := users[friendID]
		friends[i] = friend.Name
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(friends)
}

func updateAge(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data struct {
		NewAge int `json:"new_age"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Age = data.NewAge
	users[userID] = user

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User's age has been updated to %d\n", user.Age)
}
