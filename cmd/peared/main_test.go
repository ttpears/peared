package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peared/peared/internal/daemon"
)

func TestPromptAdapterSelection_Default(t *testing.T) {
	adapters := []daemon.Adapter{
		{ID: "hci0", Alias: "Host", Address: "AA:BB", Powered: true, Transport: daemon.AdapterTransportUSB},
		{ID: "hci1", Alias: "Dongle", Address: "CC:DD", Transport: daemon.AdapterTransportPCI},
	}

	input := strings.NewReader("\n")
	var out bytes.Buffer

	selected, err := promptAdapterSelection(input, &out, adapters, "hci0")
	if err != nil {
		t.Fatalf("prompt returned error: %v", err)
	}

	if selected != "hci0" {
		t.Fatalf("expected default adapter hci0, got %s", selected)
	}

	if !strings.Contains(out.String(), "Using default adapter hci0.") {
		t.Fatalf("expected confirmation of default adapter, got output: %s", out.String())
	}
}

func TestPromptAdapterSelection_ExplicitChoice(t *testing.T) {
	adapters := []daemon.Adapter{
		{ID: "hci0", Alias: "Host", Address: "AA:BB", Powered: true, Transport: daemon.AdapterTransportUSB},
		{ID: "hci1", Alias: "Dongle", Address: "CC:DD", Transport: daemon.AdapterTransportPCI},
	}

	input := strings.NewReader("2\n")
	var out bytes.Buffer

	selected, err := promptAdapterSelection(input, &out, adapters, "hci0")
	if err != nil {
		t.Fatalf("prompt returned error: %v", err)
	}

	if selected != "hci1" {
		t.Fatalf("expected adapter hci1, got %s", selected)
	}

	if !strings.Contains(out.String(), "Using adapter hci1.") {
		t.Fatalf("expected confirmation of adapter selection, got output: %s", out.String())
	}
}

func TestPromptAdapterSelection_InvalidThenValid(t *testing.T) {
	adapters := []daemon.Adapter{
		{ID: "hci0", Alias: "Host", Address: "AA:BB", Powered: true, Transport: daemon.AdapterTransportUSB},
		{ID: "hci1", Alias: "Dongle", Address: "CC:DD", Transport: daemon.AdapterTransportPCI},
	}

	input := strings.NewReader("9\n2\n")
	var out bytes.Buffer

	selected, err := promptAdapterSelection(input, &out, adapters, "")
	if err != nil {
		t.Fatalf("prompt returned error: %v", err)
	}

	if selected != "hci1" {
		t.Fatalf("expected adapter hci1 after retry, got %s", selected)
	}

	if !strings.Contains(out.String(), "Invalid selection") {
		t.Fatalf("expected invalid selection warning, got output: %s", out.String())
	}
}
