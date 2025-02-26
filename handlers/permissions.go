package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"casbin-demo/models"

	"casbin-demo/enforcer"
)

func GrantPermission(w http.ResponseWriter, r *http.Request) {
	var req models.PermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	enforcer := enforcer.GetEnforcer()

	fmt.Println("Granting permission to", req.Subject, "for", req.Object, "to", req.Action, "with effect", req.Effect)
	_, err := enforcer.AddPolicy(req.Subject, req.Object, req.Action, req.Effect)
	if err != nil {
		http.Error(w, "Failed to grant permission", http.StatusInternalServerError)
		return
	}

	err = enforcer.SavePolicy()
	if err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
