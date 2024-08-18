package andrew_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/playtechnique/andrew"
)

func TestMainCalledWithHelpOptionDisplaysHelp(t *testing.T) {
	t.Parallel()

	args := []string{"--help"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 0 {
		t.Error("Expected exit value 0, received %i", exit)
	}

	if !strings.Contains(received.String(), "Usage") {
		t.Errorf("Expected help message containing 'Usage', received %s", received)
	}
}

func TestMainCalledWithNoArgsUsesDefaults(t *testing.T) {
	t.Parallel()

	contentRoot, address, baseUrl := andrew.ParseArgs([]string{})

	if contentRoot != andrew.DefaultContentRoot {
		t.Errorf("contentroot should be %s, received %q", andrew.DefaultContentRoot, contentRoot)
	}

	if address != andrew.DefaultAddress {
		t.Errorf("address should be %s, received %q", andrew.DefaultAddress, address)
	}

	if baseUrl != andrew.DefaultBaseUrl {
		t.Errorf("baseUrl should be %s, received %q", andrew.DefaultBaseUrl, baseUrl)
	}
}

func TestMainCalledWithArgsOverridesDefaults(t *testing.T) {
	t.Parallel()

	contentRoot, address, baseUrl := andrew.ParseArgs([]string{"1", "2", "3"})

	if contentRoot != "1" {
		t.Errorf("contentroot should be %s, received %q", "1", contentRoot)
	}

	if address != "2" {
		t.Errorf("address should be %s, received %q", "2", address)
	}

	if baseUrl != "3" {
		t.Errorf("baseUrl should be %s, received %q", "3", baseUrl)
	}
}

func TestMainCalledWithInvalidAddressPanics(t *testing.T) {
	t.Parallel()
	args := []string{".", "notanipaddress"}
	nullLogger := new(bytes.Buffer)

	// No need to check whether `recover()` is nil. Just turn off the panic.
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic with invalid address, received %v", err)
		}
	}()

	andrew.Main(args, nullLogger)
}

func TestMainCalledWithCertOptionWithoutPathFails(t *testing.T) {
	t.Parallel()

	args := []string{"--cert"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 1 {
		t.Error("Expected exit value 1, received %i", exit)
	}

	if !strings.Contains(received.String(), "missing certificate path") {
		t.Errorf("Expected help message containing 'missing certificate path', received %s", received)
	}
}

func TestMainCalledWithCertOptionWithoutPrivateKeyFails(t *testing.T) {
	t.Parallel()

	args := []string{"--cert", "testdata/test-cert.crt"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 1 {
		t.Error("Expected exit value 1, received %i", exit)
	}

	if !strings.Contains(received.String(), "must be provided together") {
		t.Errorf("Expected help message containing 'must be provided together', received %s", received)
	}
}

func TestMainCalledWithPrivateKeyOptionWithoutPathFails(t *testing.T) {
	t.Parallel()

	args := []string{"--privatekey"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 1 {
		t.Error("Expected exit value 1, received %i", exit)
	}

	if !strings.Contains(received.String(), "missing private key path") {
		t.Errorf("Expected help message containing 'missing private key path', received %s", received)
	}
}

func TestMainCalledWithPrivateKeyOptionWithoutCertFails(t *testing.T) {
	t.Parallel()

	args := []string{"--privatekey", "testdata/test-cert.crt"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 1 {
		t.Error("Expected exit value 1, received %i", exit)
	}

	if !strings.Contains(received.String(), "must be provided together") {
		t.Errorf("Expected help message containing 'must be provided together', received %s", received)
	}
}
