package main

import "testing"

func TestModeAvailability(t *testing.T) {
	t.Parallel()

	t.Run("invalid query returns unavailable", func(t *testing.T) {
		availabilities, err := client.ModeAvailability("fork:insights test", "literal")
		if err != nil {
			t.Error(err)
		}
		for _, response := range availabilities {
			if response.Available == true {
				t.Errorf("expected mode %v to be unavailable", response.Mode)
			}
			if response.ReasonUnavailable == nil {
				t.Errorf("expected to receive an unavailable reason, got nil")
			}
		}
	})
}
