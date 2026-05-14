package andrew

import (
	"maps"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func TestTitleDiscoveryReturnsErrorWhenNoTitleElementInPageContent(t *testing.T) {
	_, err := titleFromHTMLTitleElement([]byte("snibble"))

	if err.Error() != "no title element found" {
		t.Fatalf("Expected error, received %v", err)
	}
}

func TestTitleElementDiscoveredWhenPresentInPageContent(t *testing.T) {
	expected := "my title"
	received, err := titleFromHTMLTitleElement([]byte("<title>" + expected + "</title>"))

	if err != nil {
		t.Fatal(err)
	}

	if received != expected {
		t.Fatal(cmp.Diff(expected, received))
	}
}

func TestGetTitleReturnsPageFileNameWhenNoTitleInDocument(t *testing.T) {
	received, err := getTitle("page title", []byte{})

	if err != nil {
		t.Fatal(err)
	}

	if received != "page title" {
		t.Fatal(cmp.Diff("", received))
	}
}

func TestOneMetaTagPopulatesATag(t *testing.T) {
	expected := map[string]string{"andrew-publish-time": "2025-03-01"}
	received, err := GetMetaElements([]byte("<meta name=andrew-publish-time content=2025-03-01>"))

	if err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(expected, received) {
		t.Fatal(cmp.Diff(expected, received))
	}
}

func TestMultipleMetaTagsPopulatedWithExpectedElements(t *testing.T) {
	expected := map[string]string{"andrew-publish-time": "2025-03-01", "andrew-roflcopter": "hippolol"}
	received, err := GetMetaElements([]byte("<meta name=andrew-publish-time content=2025-03-01> <meta name=andrew-roflcopter content=hippolol>"))

	if err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(expected, received) {
		t.Fatal(cmp.Diff(expected, received))
	}
}

func TestPageFindsIncludeFiles(t *testing.T) {
	t.Parallel()
	expected := string([]byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>

<body>
</body>
`))

	testPage := []byte(`{{ .AndrewIncludeFile }}
<body>
</body>
`)

	includeFile := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
`)

	server := Server{SiteFiles: fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: testPage,
		},
		".AndrewIncludeFile": &fstest.MapFile{
			Data: includeFile,
		},
	}}

	page, err := NewPage(server, "index.html")

	if err != nil {
		t.Error(err)
	}

	if page.Content != string(expected) {
		t.Error(cmp.Diff(expected, page.Content))
	}
}
func TestIncludeFileCanBeFoundWithNonDefaultIncludeName(t *testing.T) {
	t.Parallel()
	expected := string([]byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>

<body>
</body>
`))

	testPage := []byte(`{{ .AndrewIncludeFile.test }}
<body>
</body>
`)

	includeFile := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
`)

	server := Server{SiteFiles: fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: testPage,
		},
		".AndrewIncludeFile.test": &fstest.MapFile{
			Data: includeFile,
		},
	}}

	page, err := NewPage(server, "index.html")

	if err != nil {
		t.Fatal(err)
	}

	if page.Content != string(expected) {
		t.Errorf("--Expected:\n%s\n--Received:\n%s", expected, page.Content)
	}
}

func TestMultipleIncludeFilesCanBeFoundAndInserted(t *testing.T) {
	t.Parallel()
	expected := string([]byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>

<body>
</body>
roflcopter
`))

	testPage := []byte(`{{ .AndrewIncludeFile.test }}
<body>
</body>
{{ .AndrewIncludeFile.test2 }}
`)

	includeFile := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
`)

	server := Server{SiteFiles: fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: testPage,
		},
		".AndrewIncludeFile.test": &fstest.MapFile{
			Data: includeFile,
		},
		".AndrewIncludeFile.test2": &fstest.MapFile{
			Data: []byte("roflcopter"),
		},
	}}

	page, err := NewPage(server, "index.html")

	if err != nil {
		t.Fatal(err)
	}

	if page.Content != string(expected) {
		t.Errorf("--Expected:\n%s\n--Received:\n%s", expected, page.Content)
	}
}

// func TestIncludeFileCanRenderVariables(t *testing.T) {
// 	t.Parallel()
// 	expected := string([]byte(`
// <!DOCTYPE html>
// <head>
//   <title>index title</title>
//   <meta name="roflcopter" content="soisoi" />
// </head>

// <body>
// </body>
// `))

// 	testPage := []byte(`{{ .AndrewIncludeFile meta-name="roflcopter" content="soisoi"}}
// <body>
// </body>
// `)

// 	includeFile := []byte(`
// <!DOCTYPE html>
// <head>
//   <title>index title</title>
// </head>
// `)

// 	server := Server{SiteFiles: fstest.MapFS{
// 		"index.html": &fstest.MapFile{
// 			Data: testPage,
// 		},
// 		".AndrewIncludeFile": &fstest.MapFile{
// 			Data: includeFile,
// 		},
// 	}}

// 	page, err := NewPage(server, "index.html")

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if page.Content != string(expected) {
// 		t.Errorf("Expected:\n%s\nReceived:\n%s", expected, page.Content)
// 	}
// }

func TestIncludeREPattern(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMatch   bool
		wantCapture string
	}{
		{
			name:        "matches basic include",
			input:       "{{ .AndrewIncludeFile }}",
			wantMatch:   true,
			wantCapture: ".AndrewIncludeFile",
		},
		{
			name:        "matches include with single extension",
			input:       "{{ .AndrewIncludeFile.test }}",
			wantMatch:   true,
			wantCapture: ".AndrewIncludeFile.test",
		},
		{
			name:        "matches include with multiple extensions",
			input:       "{{ .AndrewIncludeFile.test.foo }}",
			wantMatch:   true,
			wantCapture: ".AndrewIncludeFile.test.foo",
		},
		{
			name:        "does not match without spaces",
			input:       "{{.AndrewIncludeFile}}",
			wantMatch:   false,
			wantCapture: "",
		},
		{
			name:        "does not match with extra spaces",
			input:       "{{  .AndrewIncludeFile }}",
			wantMatch:   false,
			wantCapture: "",
		},
		{
			name:        "does not match partial name",
			input:       "{{ .AndrewInclude }}",
			wantMatch:   false,
			wantCapture: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := includeRE.FindStringSubmatch(tt.input)

			gotMatch := matches != nil
			if gotMatch != tt.wantMatch {
				t.Errorf("match = %v, want %v", gotMatch, tt.wantMatch)
			}

			if tt.wantMatch && matches != nil {
				captureIdx := includeRE.SubexpIndex(andrewIncludeFileCaptureGroup)
				gotCapture := matches[captureIdx]
				if gotCapture != tt.wantCapture {
					t.Errorf("captured = %q, want %q", gotCapture, tt.wantCapture)
				}
			}
		})
	}
}
