package apis

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
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}

	var userRequest UserRequest
	err = json.Unmarshal(body, &userRequest)
	if err != nil {
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}

	user := userRequest.User
	fmt.Println("user params:", user)

	if user.Uid == "" {
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
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
			models.WriteServiceError(w, "Failed to save user", false, false, http.StatusInternalServerError)
			if err != nil {
				log.Printf("Failed to write service error: %v", err)
			}
			return
		}
	}

	result[0] = user

	// return user data
	// service response
	models.WriteServiceResponse(w, "Success", result, true, true, http.StatusOK)
	fmt.Println("---User response sent---")
}
