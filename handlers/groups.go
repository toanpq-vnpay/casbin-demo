package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"casbin-demo/enforcer"

	"github.com/gorilla/mux"
)

func AddUserToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	group := vars["groupname"]

	enforcer := enforcer.GetEnforcer()
	fmt.Println("Adding user", username, "to group", group)
	_, err := enforcer.AddGroupingPolicy(username, group)
	if err != nil {
		http.Error(w, "Failed to add user to group", http.StatusInternalServerError)
		return
	}

	err = enforcer.SavePolicy()
	if err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveUserFromGroup removes a user from a specific group
func RemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	groupname := vars["groupname"]

	e := enforcer.GetEnforcer()

	fmt.Println("Removing user", username, "from group", groupname)
	// Remove user from group using Casbin's API
	removed, err := e.DeleteRoleForUser(username, groupname)
	if err != nil {
		http.Error(w, "Failed to remove user from group", http.StatusInternalServerError)
		return
	}

	if !removed {
		http.Error(w, "User is not in the specified group", http.StatusNotFound)
		return
	}

	// Save the policy after removing the user
	err = e.SavePolicy()
	if err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User successfully removed from group",
	})
}

// GetGroupUsers retrieves all users belonging to a specific group
func GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupname := vars["groupname"]

	e := enforcer.GetEnforcer()

	// Get all users in the group using Casbin's API
	users, err := e.GetImplicitUsersForRole(groupname)
	if err != nil {
		http.Error(w, "Failed to get group users", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]string{})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// DeleteGroup deletes a group and all its associated permissions
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupname := vars["groupname"]

	e := enforcer.GetEnforcer()

	// Delete the role (group)
	removed, err := e.DeleteRole(groupname)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	if !removed {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Save the policy after removing the group
	err = e.SavePolicy()
	if err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group successfully deleted",
	})
}

// DeletePermissions deletes all permissions for a user or group
func DeletePermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	e := enforcer.GetEnforcer()

	// Remove all permissions for the user/group
	removed, err := e.DeletePermissionsForUser(name)
	if err != nil {
		http.Error(w, "Failed to delete permissions", http.StatusInternalServerError)
		return
	}

	if !removed {
		http.Error(w, "No permissions found for the specified name", http.StatusNotFound)
		return
	}

	// Save the policy after removing the permissions
	err = e.SavePolicy()
	if err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Permissions successfully deleted",
	})
}
