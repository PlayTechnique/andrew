package andrew

import (
	"maps"
	"testing"
	"testing/fstest"
	"time"

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

func TestPageFindsPartials(t *testing.T) {
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

	partialFile := []byte(`
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
			Data: partialFile,
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
func TestPartialCanBeFoundWithNonDefaultName(t *testing.T) {
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

	partialFile := []byte(`
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
			Data: partialFile,
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

func TestMultiplePartialsCanBeFoundAndInserted(t *testing.T) {
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

	partialFile := []byte(`
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
			Data: partialFile,
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

// func TestpartialFileCanRenderVariables(t *testing.T) {
// 	t.Parallel()

// 	testPage := []byte(`{{ .AndrewPartialFile meta-name="roflcopter" content="soisoi"}}
// <body>
// </body>
// `)

// 	partialFile := []byte(`
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
// 			Data: partialFile,
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
func TestPartialREPattern(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMatch   bool
		wantCapture string
	}{
		{
			name:        "matches basic partial",
			input:       "{{ .AndrewPartialFile }}",
			wantMatch:   true,
			wantCapture: ".AndrewPartialFile",
		},
		{
			name:        "matches partial with single extension",
			input:       "{{ .AndrewPartialFile.test }}",
			wantMatch:   true,
			wantCapture: ".AndrewPartialFile.test",
		},
		{
			name:        "matches partial with multiple extensions",
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
			name:        "does not match incomplete name",
			input:       "{{ .AndrewPartial }}",
			wantMatch:   false,
			wantCapture: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := partialParser()
			matches := parser.regex.FindStringSubmatch(tt.input)

			gotMatch := matches != nil
			if gotMatch != tt.wantMatch {
				t.Errorf("match = %v, want %v", gotMatch, tt.wantMatch)
			}

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
			res := parsePartialDataTags(tt.dataTags)

			if !maps.Equal(res, tt.want) {
				t.Errorf("received = %v || want %v", res, tt.want)
			}

		})
	}
}
func TestPartialDataIsParsed(t *testing.T) {
	tests := []struct {
		name         string
		testPage     []byte
		partialFiles map[string][]byte
		expected     string
	}{
		{
			name:     "partial with single data attribute",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true\" }}\n"),
			partialFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>Name: {{ .metaname }}</p>"),
			},
			expected: "<p>Name: true</p>\n",
		},
		{
			name:     "partial with multiple data attributes",
			testPage: []byte("{{ .AndrewPartialFile metaname='Bob' metadate=\"2006-03-04\" }}\n"),
			partialFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>{{ .metaname }} on '{{ .metadate }}'</p>"),
			},
			expected: "<p>'Bob' on '2006-03-04'</p>\n",
		},
		{
			name:     "last value wins when key repeated",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true\" metaname=\"false\" }}\n"),
			partialFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>{{ .metaname }}</p>"),
			},
			expected: "<p>false</p>\n",
		},
		{
			name:     "partial files provided with data tags that don't include anywhere do not blow up the parser",
			testPage: []byte("{{ .AndrewPartialFile metaname=true }}\n"),
			partialFiles: map[string][]byte{
				".AndrewPartialFile": []byte("<p>Static content</p>"),
			},
			expected: "<p>Static content</p>\n",
		},
		{
			name:     "Values can have spaces",
			testPage: []byte("{{ .AndrewPartialFile metaname=\"true beans\" }}\n"),
			partialFiles: map[string][]byte{
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

			for path, content := range tt.partialFiles {
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

func TestPagesInDir(t *testing.T) {
	tests := []struct {
		name         string
		startDir     string
		wantUrlPaths []string
	}{
		{
			name:     "root walks the whole site, skipping index and non-html files",
			startDir: ".",
			wantUrlPaths: []string{
				"blog/newest.html",
				"blog/oldest.html",
				"blog/reindex.html",
				"blog/untitled.html",
				"page.html",
			},
		},
		{
			name:     "a subdirectory scopes the walk to that directory",
			startDir: "blog",
			wantUrlPaths: []string{
				"blog/newest.html",
				"blog/oldest.html",
				"blog/reindex.html",
				"blog/untitled.html",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages, err := pagesInDir(testSiteFiles(), tt.startDir)
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for _, page := range pages {
				got = append(got, page.UrlPath)
			}

			if diff := cmp.Diff(tt.wantUrlPaths, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestPagesInDirExtractsTitleAndPublishTime(t *testing.T) {
	pages, err := pagesInDir(testSiteFiles(), "blog")
	if err != nil {
		t.Fatal(err)
	}

	byPath := map[string]Page{}
	for _, page := range pages {
		byPath[page.UrlPath] = page
	}

	tests := []struct {
		name            string
		urlPath         string
		wantTitle       string
		wantPublishTime time.Time
	}{
		{
			name:            "explicit title element and andrew-publish-time meta",
			urlPath:         "blog/newest.html",
			wantTitle:       "Newest",
			wantPublishTime: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:            "no title element falls back to basename, no meta falls back to ModTime",
			urlPath:         "blog/untitled.html",
			wantTitle:       "untitled.html",
			wantPublishTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, ok := byPath[tt.urlPath]
			if !ok {
				t.Fatalf("%s missing from returned pages", tt.urlPath)
			}
			if page.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", page.Title, tt.wantTitle)
			}
			if !page.PublishTime.Equal(tt.wantPublishTime) {
				t.Errorf("PublishTime = %s, want %s", page.PublishTime, tt.wantPublishTime)
			}
		})
	}
}

func TestPagesInDirErrorsOnMissingStartDir(t *testing.T) {
	_, err := pagesInDir(testSiteFiles(), "does-not-exist")
	if err == nil {
		t.Fatal("expected an error for a startDir that is not in the fs.FS, got nil")
	}
}

// TestPagesInDirSkipsPagesWithBrokenPartials pins the choice to degrade one page rather than
// fail everything around it: an author who typos a partial name should lose that page from
// the listings, not the whole rss feed or table of contents that the page appears in.
func TestPagesInDirSkipsPagesWithBrokenPartials(t *testing.T) {
	siteFiles := fstest.MapFS{
		"good.html": &fstest.MapFile{Data: []byte("<title>Good</title>")},
		// No .AndrewPartialFile.missing exists anywhere, so findPartialFile walks to the
		// root of the fs and gives up.
		"broken.html": &fstest.MapFile{Data: []byte("<title>Broken</title>{{ .AndrewPartialFile.missing }}")},
	}

	pages, err := pagesInDir(siteFiles, ".")
	if err != nil {
		t.Fatal(err)
	}

	got := []string{}
	for _, page := range pages {
		got = append(got, page.UrlPath)
	}

	if diff := cmp.Diff([]string{"good.html"}, got); diff != "" {
		t.Error(diff)
	}
}

// testSiteFiles is a fixture site reflecting the basic cases pagesInDir has to handle:
// index.html exclusion, non-html exclusion, directory scoping, both title sources and both
// publish-time sources. Every page gets a distinct date so sort order is unambiguous.
func testSiteFiles() fstest.MapFS {
	return fstest.MapFS{
		// Excluded: index pages are never listed.
		"index.html":      &fstest.MapFile{Data: []byte("<title>Home</title>")},
		"blog/index.html": &fstest.MapFile{Data: []byte("<title>Blog</title>")},

		// Excluded: not html. htmlguide.txt has "html" in its name, and html/notes.txt
		// lives in a directory called html, but neither is an html page.
		"styles.css":     &fstest.MapFile{Data: []byte("body {}")},
		"blog/notes.txt": &fstest.MapFile{Data: []byte("scratch")},
		"htmlguide.txt":  &fstest.MapFile{Data: []byte("prose about html")},
		"html/notes.txt": &fstest.MapFile{Data: []byte("more prose")},

		// Included: explicit <title> and explicit andrew-publish-time.
		"page.html": &fstest.MapFile{Data: []byte(
			`<head><title>Root Page</title><meta name="andrew-publish-time" content="2024-01-01"></head>`)},

		// Included: the filename contains "index.html" as a substring, but it is a real
		// page rather than an index.
		"blog/reindex.html": &fstest.MapFile{Data: []byte(
			`<head><title>Reindex</title><meta name="andrew-publish-time" content="2023-06-01"></head>`)},
		"blog/newest.html": &fstest.MapFile{Data: []byte(
			`<head><title>Newest</title><meta name="andrew-publish-time" content="2024-03-01"></head>`)},
		"blog/oldest.html": &fstest.MapFile{Data: []byte(
			`<head><title>Oldest</title><meta name="andrew-publish-time" content="2024-02-01"></head>`)},

		// Included: no <title> -> title falls back to basename;
		// no meta -> publish time falls back to ModTime.
		"blog/untitled.html": &fstest.MapFile{
			Data:    []byte("<p>no title element here</p>"),
			ModTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}
