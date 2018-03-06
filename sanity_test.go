package main

import (
	"os/exec"
	"testing"
)

func TestRunGoVet(t *testing.T) {
	vet := exec.Command("go", "vet")
	err := vet.Run()

	if err != nil {
		t.Error(err)
	}
}

func TestRunGoFmt(t *testing.T) {
	fmtCmd := exec.Command("go", "fmt")
	err := fmtCmd.Run()

	if err != nil {
		t.Error(err)
	}
}
