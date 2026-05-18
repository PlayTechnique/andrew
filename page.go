package andrew

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/net/html"
)

// A regular expression match returns as {"parentKey": {"key", "value"}}.
// I want to anchor everything here and in tests on the AndrewPartialFile parent key.
type includeParser struct {
	fileParentKey string
	dataParentKey string
	regex         *regexp.Regexp
}

var (
	parserInstance *includeParser
	parserOnce     sync.Once
)

// getIncludeParser returns the singleton includeParser instance.
func getIncludeParser() *includeParser {
	parserOnce.Do(func() {
		parserInstance = &includeParser{
			fileParentKey: "AndrewPartialFile",
			dataParentKey: "AndrewPartialFileData",
		}
		parserInstance.regex = regexp.MustCompile(
			fmt.Sprintf(`{{ (?P<%s>\.AndrewPartialFile[\w.]*)(?P<%s>.*?)\s*?}}`,
				parserInstance.fileParentKey,
				parserInstance.dataParentKey))
	})
	return parserInstance
}

// Page tracks the content of a specific file and various pieces of metadata about it.
// The Page makes creating links and serving content convenient, as it lets me offload
// the parsing of any elements into a constructor, so that when I need to present those
// elements to an end-user they're easy for me to reason about.
type Page struct {
	// Page title
	Title string
	// According to https://datatracker.ietf.org/doc/html/rfc1738#section-3.1, the subsection of a
	// URL after the protocol://hostname is the UrlPath.
	UrlPath     string
	Content     string
	PublishTime time.Time
}

type TagInfo struct {
	Data       string
	Attributes map[string]string
}

// NewPage creates a Page from a URL by reading the corresponding file from the
// Server's SiteFiles.
// NewPage does this by reading the page content from disk, then parsing out various
// metadata that are convenient to have quick access to, such as the page title or the
// publish time.
func (s Server) NewPage(pageUrl string) (Page, error) {
	pageContent, err := fs.ReadFile(s.SiteFiles, pageUrl)
	if err != nil {
		return Page{}, err
	}

	// The fs.FS documentation notes that paths should not start with a leading slash.
	pagePath := strings.TrimPrefix(pageUrl, "/")

	pageTitle, err := getTitle(pagePath, pageContent)
	if err != nil {
		return Page{}, err
	}

	pagePublishTime, err := getPublishTime(s.SiteFiles, pagePath, pageContent)

	if err != nil {
		return Page{}, err
	}

	renderedPageContent, err := renderIncludeFiles(s.SiteFiles, pagePath, pageContent)
	if err != nil {
		return Page{}, err
	}

	page := Page{Content: string(renderedPageContent), PublishTime: pagePublishTime, Title: pageTitle, UrlPath: pageUrl}

	siblings, err := s.GetSiblingsAndChildren(page.UrlPath)

	if err != nil {
		return page, err
	}

	orderedSiblings := SortPagesByDate(siblings)

	// Only execute templates for html files, not pngs or other kinds of file.
	// This is so the template rendering engine doesn't receive a binary blob, which
	// makes it panic.
	if strings.HasSuffix(page.UrlPath, ".html") {
		contentWithContents, err := RenderTableOfContents(orderedSiblings, page)
		if err != nil {
			return Page{}, err
		}

		page.Content = string(contentWithContents)
	}

	return page, nil
}

func getPublishTime(siteFiles fs.FS, pagePath string, pageContent []byte) (time.Time, error) {
	pageInfo, err := fs.Stat(siteFiles, pagePath)
	if err != nil {
		return time.Time{}, err
	}

	publishTime := pageInfo.ModTime()

	meta, err := GetMetaElements(pageContent)
	if err != nil {
		return publishTime, err
	}

	metaPublishTime, ok := meta["andrew-publish-time"]

	if ok {
		andrewCreatedAt, err := time.Parse(time.DateTime, metaPublishTime)

		// Check if the error is of type *time.ParseError as this indicates
		// we may have no timestamp with the date
		if _, ok := err.(*time.ParseError); ok {
			andrewCreatedAt, err = time.Parse(time.DateOnly, metaPublishTime)
		}

		// The errors that come out of time.Parse are all not interesting to me; we just want
		// to use those errors to tell us if it's safe to set PublishTime to the value of the
		// meta element.
		if err == nil {
			publishTime = andrewCreatedAt
		}
	}

	return publishTime, nil
}

// SetUrlPath updates the UrlPath on a pre-existing Page.
func SetUrlPath(page Page, urlPath string) Page {
	page.UrlPath = urlPath
	return page
}

// getTagInfo recursively descends an html node tree for the requested tag,
// searching both data and attributes to find information about the node that's requested.
func getTagInfo(tag string, n *html.Node) TagInfo {
	var tagDataAndAttributes TagInfo = TagInfo{
		Data:       "",
		Attributes: make(map[string]string),
	}

	// getTag recursively descends an html node tree, searching for
	// the attribute provided. Once the attribute is discovered, it first checks
	// for any Attributes available on the html node. If there are no Attributes,
	// the key won't exist in the tagDataAndAttributes map.
	// If there is data, it will append to attributes.
	var getTag func(n *html.Node)

	getTag = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			attrName := ""
			attrVal := ""

			if n.Attr != nil {
				for _, attr := range n.Attr {
					switch attr.Key {
					case "content":
						attrVal = attr.Val
					case "name":
						attrName = attr.Val
					}
					tagDataAndAttributes.Attributes[attrName] = attrVal
				}
			}

			if n.FirstChild != nil {
				tagDataAndAttributes.Data = n.FirstChild.Data
			}

			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getTag(c)
		}
	}

	// Start the recursion from the root node
	getTag(n)

	return tagDataAndAttributes
}

func GetMetaElements(htmlContent []byte) (map[string]string, error) {
	element := "meta"

	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return map[string]string{}, err
	}

	tagInfo := getTagInfo(element, doc)

	return tagInfo.Attributes, nil
}

func getTitle(htmlFilePath string, htmlContent []byte) (string, error) {
	title, err := titleFromHTMLTitleElement(htmlContent)
	if err != nil {
		if err.Error() != "no title element found" {
			return "", err
		}
		// filename is bam.html
		title = path.Base(htmlFilePath)
	}
	return title, nil
}

//	renderIncludeFiles parses the syntax {{ .AndrewPartialFile foo=bar bam=bas }}, finds the include
//
// file on the file system, reads its contents and performs a template.Execute against it using the key/value pairs as a hashmap.
// params:
//  1. siteFiles: We need this to be able to find the include files.
//  2. pagePath: This is the path to the currently-being-evaluated page. Finding includes begins in this page's directory and heads
//     upwards to find the include file.
//  3. pageContent: The page's contents; the include statement will be parsed out of these.
//
// retval
// []byte: an array of bytes representing the new version of pageContent, with the includes included.
// error: as normal.
func renderIncludeFiles(siteFiles fs.FS, pagePath string, pageContent []byte) ([]byte, error) {
	// The parser will parse pageContent for the include statements.
	includeParser := getIncludeParser()
	matches := includeParser.regex.FindAllStringSubmatch(string(pageContent), -1)

	// no .AndrewPartialFile directive to parse. Good to bail.
	if matches == nil {
		slog.Debug("renderIncludeFile", "AndrewIncludeDirectiveFound", "false", "pageContent", pageContent)
		return pageContent, nil
	}

	// Each match in matches has 2 components:
	// match[0] == the entire matched string.
	// match[n > 0] == the contents of the Nth capture group.
	fileIndex := includeParser.regex.SubexpIndex("AndrewPartialFile")     // returns 1
	dataIndex := includeParser.regex.SubexpIndex("AndrewPartialFileData") // returns 2

	var templateBuffer bytes.Buffer
	var includeContent string

	for _, m := range matches {
		includeToFind := m[fileIndex]
		// It's pretty common to put ' and " in the value of the k/v pair in the include statement.
		// Cleaning them up is a better experience than failing the parsing.
		dataTagsToParse := m[dataIndex]

		// We always need to know the path to the required include file, so that we
		// can read in the include file to insert it into the web page in place of the {{ }} statement.
		includeFile, err := findIncludeFile(siteFiles, pagePath, includeToFind)

		if err != nil {
			slog.Debug("renderIncludeFile renderFile not found", "error", err)

			return pageContent, err
		}

		// Read the partial (always needed)
		partial, err := fs.ReadFile(siteFiles, includeFile)
		if err != nil {
			return pageContent, err
		}

		// The tag format {{ .AndrewPartialFile spaceman=david }} requires parsing out the key/value pairs.
		// parseIncludeDataTags returns an empty map if there's no data, which is fine for template execution.
		tags := parseIncludeDataTags(dataTagsToParse)

		// Always execute template - works with empty tags (just returns raw content)
		partialTemplate, err := template.New(includeParser.dataParentKey).Parse(string(partial))
		if err != nil {
			panic(err)
		}

		templateBuffer.Reset() // Clear buffer for reuse in loop
		err = partialTemplate.Execute(&templateBuffer, tags)
		if err != nil {
			return templateBuffer.Bytes(), err
		}

		includeContent = templateBuffer.String()

		slog.Debug("renderIncludeFiles", "includeContent", includeContent)
		pageContent = []byte(strings.Replace(string(pageContent), m[0], string(includeContent), -1))
	}

	return pageContent, nil
}

func parseIncludeDataTags(data string) map[string]string {
	slog.Debug("parseIncludeDataTags", "inputData", data)
	var tags = make(map[string]string)

	if data == "" {
		return map[string]string{"": ""}
	}

	keywordIdentified := false
	quotedValue := false
	var keyword string
	var value string

	var tok rune

	for i := 0; i < len(data); i++ {
		tok = rune(data[i])

		// Any whitespace before the keyword should be skipped
		//  skip whitespace before the next key.
		if !keywordIdentified && tok == ' ' {
			continue
		}

		switch {
		case (!keywordIdentified):
			keyword = keyword + string(tok)
			if i+1 < len(data) && data[i+1] == '=' {
				// Our keyword is identified! What if it has whitespace between its end and equals,
				// because someone wrote foo = bar?
				keyword = strings.TrimRight(keyword, " ")
				// Let's head up to the equals sign
				i = i + 1
				keywordIdentified = true

				// What if someone wrote foo = bar? Then we need to go past the whitespace after the equals sign
				for i+1 < len(data) && data[i+1] == ' ' {
					i = i + 1
				}

				// we are either at the start of a quoted value or a bare value now.
				// Before we go further in parsing, let's figure out if our value is quoted or not.
				if i+1 < len(data) && data[i+1] == '"' {
					quotedValue = true
					i = i + 1 // now we are at the start of a quoted value
				}
			}
			continue
		case (keywordIdentified):
			// For a value, if the next character is a space:
			// If we are not in a qouted value, this indicates that we are done parsing, so
			// assign the value to the keyword and proceed.
			// we are in a quotedValue. We have to check the current character is a " and the next character
			if quotedValue {
				// If the next token is a quote, we're done parsing the value!
				// We drop the quote and just put the value in the keyword
				if tok == '"' {
					keywordIdentified = false
					quotedValue = false

					tags[keyword] = value
					keyword = ""
					value = ""
					// If the next character's a space (it should be, but let's check)
					if i+1 < len(data) && data[i+1] == ' ' {
						i = i + 1 // move our marching pointer to the space, so that i++ can march past it at the start of the next iteration.
					}

					continue
				}

				value = value + string(tok)
			} else {
				// Unquoted value: a following space ends it.
				if i+1 < len(data) && data[i+1] == ' ' {
					value = value + string(tok)
					tags[keyword] = value
					keyword = ""
					value = ""
					keywordIdentified = false
					i = i + 1
					continue
				}
				value = value + string(tok)
			}
		}
	}

	// Final flush: handles `foo=bar` with no trailing whitespace.
	if keywordIdentified {
		tags[keyword] = value
	}

	return tags
}

// findIncludeFile will upwards walk from the directory containing includeName upwards in the tree.
// If the include file is found in the same file as the pagePath, the directory containing the
// include file is "."; inside an fs.FS "." is only allowed to refer to the root directory, not
// the pwd, but fortunately path.Join takes care of that issue for us.
// args:
// 1. siteFiles fs.FS - a file system to search for an include file
// 2. pagePath - the path to the starting page for the search. We simply interrogate this for the containing directory path, to start the upwards walk
// 3. includeName - the name of the include file to look for
func findIncludeFile(siteFiles fs.FS, pagePath string, includeName string) (string, error) {
	pagePwd := path.Dir(pagePath)

	slog.Debug("findIncludeFile", "pagePath", pagePath)
	for {
		candidate := path.Join(pagePwd, includeName)

		_, err := fs.Stat(siteFiles, candidate)

		// If the candidate is a real file, we're done.
		if err == nil {
			slog.Debug("findIncludeFile", "foundFile", candidate)

			return candidate, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}

		// Exit out of the walk in the root dir; nowhere else to look.
		if pagePwd == "." {
			return "", fs.ErrNotExist
		}

		// reset dir upwards one level
		pagePwd = path.Dir(pagePwd)
	}
}

// titleFromHTMLTitleElement returns the content of the "title" tag or an empty string.
// The error value "no title element found" is returned if title is not discovered
// or is set to an empty string.
func titleFromHTMLTitleElement(fileContent []byte) (string, error) {

	doc, err := html.Parse(bytes.NewReader(fileContent))
	if err != nil {
		return "", err
	}

	tagInfo := getTagInfo("title", doc)
	if len(tagInfo.Data) == 0 {
		return "", fmt.Errorf("no title element found")
	}
	return tagInfo.Data, nil
}

func SortPagesByDate(pages []Page) []Page {

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].PublishTime.After(pages[j].PublishTime)
	})

	return pages
}
