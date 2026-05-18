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

	testPage := []byte(`{{ .AndrewPartialFile }}
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
		".AndrewPartialFile": &fstest.MapFile{
			Data: includeFile,
		},
	}}

	page, err := server.NewPage("index.html")

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

	testPage := []byte(`{{ .AndrewPartialFile.test }}
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
		".AndrewPartialFile.test": &fstest.MapFile{
			Data: includeFile,
		},
	}}

	page, err := server.NewPage("index.html")

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

	testPage := []byte(`{{ .AndrewPartialFile.test }}
<body>
</body>
{{ .AndrewPartialFile.test2 }}
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
		".AndrewPartialFile.test": &fstest.MapFile{
			Data: includeFile,
		},
		".AndrewPartialFile.test2": &fstest.MapFile{
			Data: []byte("roflcopter"),
		},
	}}

	page, err := server.NewPage("index.html")

	if err != nil {
		t.Fatal(err)
	}

	if page.Content != string(expected) {
		t.Errorf("--Expected:\n%s\n--Received:\n%s", expected, page.Content)
	}
}

// func TestIncludeFileCanRenderVariables(t *testing.T) {
// 	t.Parallel()

// 	testPage := []byte(`{{ .AndrewPartialFile meta-name="roflcopter" content="soisoi"}}
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
// 		".AndrewPartialFile": &fstest.MapFile{
// 			Data: includeFile,
// 		},
// 	}}

// 	page, err := server.NewPage("index.html")

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if page.Content != string(expected) {
// 		t.Errorf("Expected:\n%s\nReceived:\n%s", expected, page.Content)
// 	}
// }

// Verify that the regular expression used for finding Partials is working well.
// Pulling these into their own test is completely worth it; the integration style
// tests don't get this specific easily.
func TestIncludeREPattern(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMatch   bool
		wantCapture string
	}{
		{
			name:        "matches basic include",
			input:       "{{ .AndrewPartialFile }}",
			wantMatch:   true,
			wantCapture: ".AndrewPartialFile",
		},
		{
			name:        "matches include with single extension",
			input:       "{{ .AndrewPartialFile.test }}",
			wantMatch:   true,
			wantCapture: ".AndrewPartialFile.test",
		},
		{
			name:        "matches include with multiple extensions",
			input:       "{{ .AndrewPartialFile.test.foo }}",
			wantMatch:   true,
			wantCapture: ".AndrewPartialFile.test.foo",
		},
		{
			name:        "does not match without spaces",
			input:       "{{.AndrewPartialFile}}",
			wantMatch:   false,
			wantCapture: "",
		},
		{
			name:        "does not match with extra spaces",
			input:       "{{  .AndrewPartialFile }}",
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
			parser := getIncludeParser()
			matches := parser.regex.FindStringSubmatch(tt.input)

			gotMatch := matches != nil
			if gotMatch != tt.wantMatch {
				t.Errorf("match = %v, want %v", gotMatch, tt.wantMatch)
			}

			// It's right below that we actually test something. This test is why
			// includeRE and andrewIncludeFileCaptureGroup are available outside of a specific function.
			if tt.wantMatch && matches != nil {
				captureIndex := parser.regex.SubexpIndex(parser.fileParentKey)
				results := matches[captureIndex]

				if results != tt.wantCapture {
					t.Errorf("captured = %q, want %q", results, tt.wantCapture)
				}
			}
		})
	}
}

func TestDataTagParsing(t *testing.T) {
	tests := []struct {
		name        string
		dataTags    string
		shouldError bool
		want        map[string]string
	}{
		{
			name:     "empty string parses to nil map",
			dataTags: "",
			want:     map[string]string{"": ""},
		},
		{
			name:     "Well-formed pair is parsed",
			dataTags: "meta-name=roflcopter",
			want:     map[string]string{"meta-name": "roflcopter"},
		},
		{
			name:     "Key with no value returns empy",
			dataTags: "meta-name=",
			want:     map[string]string{"meta-name": ""},
		},
		{
			name:     "Value with no key returns empty",
			dataTags: "=roflcopter",
			want:     map[string]string{},
		},
		{
			name:     "Multiple pairs are parsed",
			dataTags: "meta-name=roflcopter meta-date=hippololamus",
			want:     map[string]string{"meta-name": "roflcopter", "meta-date": "hippololamus"},
		},
		{
			name:     "Whitespace is ignored",
			dataTags: "    meta-name=roflcopter  meta-date=hippololamus   ",
			want:     map[string]string{"meta-name": "roflcopter", "meta-date": "hippololamus"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseIncludeDataTags(tt.dataTags)

			if !maps.Equal(res, tt.want) {
				t.Errorf("received = %v || want %v", res, tt.want)
			}

		})
	}
}
func TestIncludePatternCapturesData(t *testing.T) {
	tests := []struct {
		name         string
		testPage     []byte
		includeFiles map[string][]byte
		expected     string
	}{
		{
			name:     "include with single data attribute",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true\" }}\n"),
			includeFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>Name: {{ .metaname }}</p>"),
			},
			expected: "<p>Name: true</p>\n",
		},
		{
			name:     "include with multiple data attributes",
			testPage: []byte("{{ .AndrewPartialFile metaname='Bob' metadate=\"2006-03-04\" }}\n"),
			includeFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>{{ .metaname }} on '{{ .metadate }}'</p>"),
			},
			expected: "<p>'Bob' on '2006-03-04'</p>\n",
		},
		{
			name:     "last value wins when key repeated",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true\" metaname=\"false\" }}\n"),
			includeFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>{{ .metaname }}</p>"),
			},
			expected: "<p>false</p>\n",
		},
		{
			name:     "include files provided with data tags that don't include anywhere doesn't blow up the parser",
			testPage: []byte("{{ .AndrewPartialFile metaname=true }}\n"),
			includeFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>Static content</p>"),
			},
			expected: "<p>Static content</p>\n",
		},
		{
			name:     "Values can have spaces",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true beans\" }}\n"),
			includeFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>{{ .metaname }}</p>"),
			},
			expected: "<p>true beans</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapFS := fstest.MapFS{
				"index.html": &fstest.MapFile{
					Data: tt.testPage,
				},
			}

			for path, content := range tt.includeFiles {
				mapFS[path] = &fstest.MapFile{
					Data: content,
				}
			}

			server := Server{SiteFiles: mapFS}

			page, err := server.NewPage("index.html")

			if err != nil {
				t.Fatal(err)
			}

			if page.Content != tt.expected {
				t.Errorf("--Expected:\n%s\n--Received:\n%s", tt.expected, page.Content)
			}
		})
	}
}
