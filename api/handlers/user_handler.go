package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/models"
)

// HandleSignIn handles POST requests to save user data
func HandleSignIn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("@HandleSignIn@2", "request is comming in")
	// Parse the request body
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// http.Error(w, "Invalid request payload", http.StatusBadRequest)
		response := models.CreateServiceResponse("Invalid request payload", 400, make([]string, 0), false)
		models.WriteServiceError[[]string](w, response, false)
		return
	}
	if user.Uid == "" {
		response := models.CreateServiceResponse("User ID is required", 400, make([]string, 0), false)
		models.WriteServiceError[[]string](w, response, false)
		return
	}
	// check if user exists
	user, err = user.CheckUser(user.Uid)
	result := make([]models.User, 1)
	if err != nil {
		user, err = user.Save()
		if err != nil {
			// http.Error(w, "Failed to save user", http.StatusInternalServerError)
			response := models.CreateServiceResponse("Failed to save user", 500, result, true)
			models.WriteServiceError[[]models.User](w, response, true)
			return
		}
	}

	// update login time
	err = user.CreateMetadata()
	if err != nil {
		// http.Error(w, "Failed to save user", http.StatusInternalServerError)
		response := models.CreateServiceResponse("Failed to save user", 500, result, true)
		models.WriteServiceError[[]models.User](w, response, true)
		return
	}

	//

	// TODO: Implement actual user saving logic (database, etc.)
	log.Printf("User received: %+v", user)

	// return user data
	// service response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.CreateServiceResponse("User saved successfully", 200, result, true)
	models.WriteServiceResponse[[]models.User](w, response, false)
}
