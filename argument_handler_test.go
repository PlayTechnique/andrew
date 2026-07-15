package andrew_test

import (
	"bytes"
	"fmt"
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

func TestMainCalledWithCertOptionWithoutPathPanics(t *testing.T) {
	t.Parallel()

	requirePanicContaining(t, "missing certificate path", func() {
		andrew.Main([]string{"--cert"}, new(bytes.Buffer))
	})
}

func TestMainCalledWithRssDirOptionWithoutPathPanics(t *testing.T) {
	t.Parallel()

	requirePanicContaining(t, "missing rss directory", func() {
		andrew.Main([]string{"--rssdir"}, new(bytes.Buffer))
	})
}

// TestMainCalledWithAnRssDirThatIsNotInTheContentRootPanics covers Main resolving the rss
// dir before it builds a server, which is what makes a typo'd --rssdir fail at startup
// rather than when someone eventually requests the feed.
func TestMainCalledWithAnRssDirThatIsNotInTheContentRootPanics(t *testing.T) {
	t.Parallel()

	requirePanicContaining(t, "must be a directory inside the content root", func() {
		andrew.Main([]string{"--rssdir", "does-not-exist", "testdata"}, new(bytes.Buffer))
	})
}

func TestMainCalledWithPrivateKeyOptionWithoutPathPanics(t *testing.T) {
	t.Parallel()

	requirePanicContaining(t, "missing private key path", func() {
		andrew.Main([]string{"--privatekey"}, new(bytes.Buffer))
	})
}
func TestMainCalledWithOneCertOptionWithoutTheOtherPanics(t *testing.T) {
	t.Parallel()

	requirePanicContaining(t, "must be provided together", func() {
		andrew.Main([]string{"--cert", "testdata/test-cert.crt"}, new(bytes.Buffer))
	})

	requirePanicContaining(t, "must be provided together", func() {
		andrew.Main([]string{"--privatekey", "testdata/test-cert.crt"}, new(bytes.Buffer))
	})
}

// requirePanicContaining fails the test unless fn panics with a value whose message
// contains want.
//
// Matching on a substring of the message is only appropriate because these particular
// errors are user-facing: the panic from a bad command line is what an end user reads on
// their terminal, so the wording is part of andrew's contract with them and is a fair
// thing to assert on. Please, do not reach for this to test an internal error, where the message
// is an implementation detail assert on the error value with errors.Is instead.
//
// A deferred recover lives in here, so it is always registered before fn runs.
// Registering a recover after the call that panics silently kills the whole test binary
// rather than failing the test.
func requirePanicContaining(t *testing.T, want string, fn func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected a panic containing %q, but nothing panicked", want)
			return
		}

		// fmt.Sprint copes with both panic(err) and panic("some string").
		if got := fmt.Sprint(r); !strings.Contains(got, want) {
			t.Errorf("expected a panic containing %q, got %q", want, got)
		}
	}()

	fn()
}
