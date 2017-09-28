package pkg

import "testing"

func TestMain(m *testing.M) { // MATCH /should call os.Exit/
	m.Run()
}
