package enforcer

import (
	"fmt"

	"github.com/casbin/casbin/v2"
)

var (
	// GlobalEnforcer is the global enforcer instance
	GlobalEnforcer *casbin.Enforcer
)

// Initialize creates a new enforcer instance
func InitializeEnforcer() error {
	var err error
	GlobalEnforcer, err = casbin.NewEnforcer("./config/pbac_model.conf", "./config/policy.csv")
	if err != nil {
		return fmt.Errorf("failed to create enforcer: %w", err)
	}

	// GlobalEnforcer.AddFunction("my_key_match", func(args ...interface{}) (interface{}, error) {
	// 	key1 := args[0].(string)
	// 	key2 := args[1].(string)
	// 	return CustomKeyMatch(key1, key2), nil
	// })

	// Load the policy from CSV file
	if err := GlobalEnforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}

	fmt.Println("Enforcer initialized successfully")
	return nil
}

// GetEnforcer returns the global enforcer instance
func GetEnforcer() *casbin.Enforcer {
	return GlobalEnforcer
}
