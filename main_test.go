package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAppHelp smoke-tests the app running. The actual code in it is not
// tested at this time, the idea is to keep the main.go code minimal.
func TestAppHelp(t *testing.T) {
	app := getApp()
	app.ErrorHandling = flag.ContinueOnError
	assert.NoError(t, app.Run([]string{"ssh-each", "-h"}))
}
