package main

import (
	"github.com/anishmgoyal/calagora-admin/bootstrap"
	_ "github.com/lib/pq"
)

func main() {
	if !bootstrap.GlobalStart() {
		panic("Failed to start server.")
	}
}
