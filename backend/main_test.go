package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"threadhelpServer/utils"
	"time"
)

func TestJsonTime(t *testing.T) {
	var now time.Time = time.Now()
	var jsonNow = utils.JSONTime(now)
	b, err := json.Marshal(jsonNow)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != fmt.Sprint(now.UnixMilli()) {
		t.Fatalf("Expected %d, got %s", now.UnixMilli(), string(b))
	}
}
