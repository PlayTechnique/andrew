package andrew

import (
	"errors"
	"fmt"
	"io"
	_ "net/http/pprof"
	"os"
)

// CertInfo tracks SSL certificate information. Andrew can optionally serve HTTPS traffic,
// but to do so it has to know how to find both the path to the certificate and to the private key.
type CertInfo struct {
	CertPath       string
	PrivateKeyPath string
}

const (
	DefaultContentRoot = "."
	DefaultAddress     = ":8080"
	DefaultBaseUrl     = "http://localhost:8080"
)

// Main is the implementation of main. It's here to get main's logic into a testable package.
func Main(args []string, printDest io.Writer) int {

	certInfo, remainingArgs, err := ParseOpts(args, printDest)
	if err != nil {
		// If we display a -h or --help flag, we helped the user and it's time to exit.
		if err.Error() == "helped" {
			return 0
		}
		fmt.Fprintln(printDest, "Error:", err)
		return 1
	}

	contentRoot, address, baseUrl := ParseArgs(remainingArgs)

	fmt.Fprintf(printDest, "Serving from %s, listening on %s, serving on %s", contentRoot, address, baseUrl)

	err = ListenAndServe(os.DirFS(contentRoot), address, baseUrl, certInfo)
	if err != nil {
		panic(err)
	}

	return 0
}

// ParseOpts parses command-line options and returns a CertInfo struct,
// remaining arguments, and an error if any.
//
// The args parameter contains the command-line arguments, and printDest
// is where the help message is written if `-h` or `--help` is specified.
//
// Supported options:
//   - -c, --cert: Path to the SSL certificate file. Must be used with `--privatekey`.
//   - -p, --privatekey: Path to the private key file. Must be used with `--cert`.
//   - -h, --help: Displays the help message and returns a specific error.
//
// If only one of `--cert` or `--privatekey` is provided, an error is returned.
//
// Returns a CertInfo struct containing the SSL certificate and key paths,
// the remaining arguments, and any error encountered.
func ParseOpts(args []string, printDest io.Writer) (*CertInfo, []string, error) {
	// Whitespace formatting here provided lovingly by eyeballing it.
	help := `Usage: andrew supports both arguments and options. Arguments are positional, options are not.
	andrew [contentRoot] [address] [baseUrl] || (-c|--cert) path/to/ssl.crt (-p|--privatekey) path/to/ssl.key (-h|--help) help message
	
	Arguments:
	  contentRoot          The root directory of your content. Defaults to '.' if not specified.
	  address              The address to bind to. Defaults to 'localhost:8080' if not specified.
				If in doubt, you probably want 0.0.0.0:<something>
	  baseUrl              The protocol://hostname for your server. Defaults to 'http://localhost:8080' 
				if not specified. Used to generate sitemap/rss feed accurately.
	
	Options:
	  -c, --cert           Path to the SSL certificate file. Must be used with --privatekey. If the certificate 
				is signed by a certificate authority, the certFile should be the concatenation of 
				the server's certificate, any intermediates, and the CA's certificate.
	  -p, --privatekey     Path to the private key file. Must be used with --cert.
	  -h, --help           Display this help message.
`

	var certPath, keyPath string
	remainingArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-c", "--cert":
			if i+1 < len(args) {
				certPath = args[i+1]
				i++ // Skip the next argument as it is the cert path

				// Check if certPath is a valid file
				if err := checkFileExists(certPath); err != nil {
					return nil, nil, fmt.Errorf("certificate %w", err)
				}
			} else {
				return nil, nil, errors.New("missing certificate path after " + arg)
			}

		case "-p", "--privatekey":
			if i+1 < len(args) {
				keyPath = args[i+1]
				i++ // Skip the next argument as it is the key path

				// Check if keyPath is a valid file
				if err := checkFileExists(keyPath); err != nil {
					return nil, nil, fmt.Errorf("private key %w", err)
				}
			} else {
				return nil, nil, errors.New("missing private key path after " + arg)
			}

		case "-h", "--help":
			fmt.Fprint(printDest, help)
			return nil, nil, errors.New("helped")
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// Validate that if one of certPath or keyPath is set, the other must be set as well
	if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
		return nil, nil, errors.New("both --cert and --privateKey must be provided together")
	}

	var cert *CertInfo
	if certPath != "" && keyPath != "" {
		cert = &CertInfo{
			CertPath:       certPath,
			PrivateKeyPath: keyPath,
		}
	}

	return cert, remainingArgs, nil
}

// ParseArgs ensures command line arguments override the default settings for a new Andrew server.
func ParseArgs(args []string) (string, string, string) {
	contentRoot := DefaultContentRoot
	address := DefaultAddress
	baseUrl := DefaultBaseUrl

	if len(args) >= 1 {
		contentRoot = args[0]
	}

	if len(args) >= 2 {
		address = args[1]
	}

	if len(args) >= 3 {
		baseUrl = args[2]
	}

	return contentRoot, address, baseUrl
}

// checkFileExists checks if the given path is a valid file and returns an error if it does not exist or is not a file.
func checkFileExists(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file does not exist: " + path)
		}
		return fmt.Errorf("unable to access file: %w", err)
	}
	if fileInfo.IsDir() {
		return errors.New("path is a directory, not a file: " + path)
	}
	return nil
}
