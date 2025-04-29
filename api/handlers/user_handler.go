package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

type UserRequest struct {
	User models.User `json:"user"`
}

// HandleSignIn handles POST requests to save user data
func HandleSignIn(w http.ResponseWriter, r *http.Request) {

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	result := make([]models.User, 1)

	if err != nil {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
		return
	}

	var userRequest UserRequest
	err = json.Unmarshal(body, &userRequest)
	if err != nil {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
		return
	}

	user := userRequest.User
	fmt.Println("user params:", user)

	if user.Uid == "" {
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}
	// check if user exists
	isSaved := user.IsUserSaved(user.Uid)

	if isSaved {
		// update login time
		user.LoginAt = time.Now()
		_, err = user.Update(user.Uid)
		if err != nil {
			fmt.Println("err:", err)
		}

	} else {

		// save user
		user, err = user.Save()
		if err != nil {
			// http.Error(w, "Failed to save user", http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(nil)
			if err != nil {
				log.Printf("Failed to write service error: %v", err)
			}
			return
		}
	}

	result[0] = user

	// return user data
	// service response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Printf("Failed to write service response: %v", err)
	}
	fmt.Println("---User response sent---")
}
