package gochimp

import (
	"testing"
)

func TestMandrillSubaccounts(t *testing.T) {
	if r, e := mandrill.SubaccountAdd("test-id", "Test Name", "Test Notes", 0); e != nil {
		t.Fatal("Expected success: " + e.Error())

	} else {
		if r.Id != "test-id" {
			t.Error("Expected id 'test-id'")
		}
		if r.Name != "Test Name" {
			t.Error("Expected name 'Test Name'")
		}
	}

	if _, e := mandrill.SubaccountInfo("test-id"); e != nil {
		t.Fatal("Expected success: " + e.Error())
	}

	if _, e := mandrill.SubaccountPause("test-id"); e != nil {
		t.Fatal("Expected success: " + e.Error())
	}

	if _, e := mandrill.SubaccountResume("test-id"); e != nil {
		t.Fatal("Expected success: " + e.Error())
	}

	if _, e := mandrill.SubaccountDelete("test-id"); e != nil {
		t.Fatal("Expected success: " + e.Error())
	}

	if _, e := mandrill.SubaccountInfo("test-id"); e == nil {
		t.Fatal("Expected failure: " + e.Error())
	}
}
