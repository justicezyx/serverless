package core

import (
	"testing"
)

func TestAllowUserAPI(t *testing.T) {
	uam := NewUserAPIManager()

	// Allow a user to access an API
	uam.AllowUserAPI("user1", "api1")

	if !uam.IsUserAllowed("user1", "api1") {
		t.Errorf("Expected user1 to be allowed to access api1")
	}
}

func TestIsUserAllowed(t *testing.T) {
	uam := NewUserAPIManager()

	// Initially, the user should not be allowed to access the API
	if uam.IsUserAllowed("user1", "api1") {
		t.Errorf("Expected user1 to not be allowed to access api1 initially")
	}

	// Allow the user to access the API
	uam.AllowUserAPI("user1", "api1")

	if !uam.IsUserAllowed("user1", "api1") {
		t.Errorf("Expected user1 to be allowed to access api1 after granting permission")
	}

	// Check another user that hasn't been granted access
	if uam.IsUserAllowed("user2", "api1") {
		t.Errorf("Expected user2 to not be allowed to access api1")
	}

	// Check another API for the same user
	if uam.IsUserAllowed("user1", "api2") {
		t.Errorf("Expected user1 to not be allowed to access api2")
	}
}
