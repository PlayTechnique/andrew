# andrew
Andrew is a web server like 90s web servers used to be™. It renders web pages from the file system, no databases
involved. This is the basic design restriction that informs feature decisions. You get started by writing an "index.html"
file and then running Andrew from that directory. 

I wanted an http server that allows me to add a simple go template instruction into an index.html that is replaced
with the contents of any html files that are below the current index.html in the file system hierarchy.

Andrew contains the concept of an AndrewPage. This structure makes various pieces of metadata stored within your web page
available to Andrew for creating links and sorting pages in the various tables of contents available (see below). The specifics
are explained below, but conceptually I'm trying to use standard html elements to inform Andrew about site metadata. For more 
you may want to check the [Architecture.md](./ARCHITECTURE.md)

Andrew includes a simple sitemap generator. Your new website needs some way to establish its identity with search engines,
after all.

## To install it

`go install github.com/playtechnique/andrew/cmd/andrew`

## invocation
andrew -h to see the help

andrew accepts up to three arguments, in this order:
```bash
andrew [contentRoot] [address] [baseUrl]
```
contentRoot is the directory you're serving from, that contains your top level index.html. andrew follows
apache's lead on expecting index.html in any directory as a default page.

address is the address you want to bind the server to. Specify as an address:port combination.

baseUrl is the hostname you're serving from. This is a part of sitemaps and future rss feeds. It also contains the protocol
e.g. `https://playtechnique.io`

It, unfortunately, does not terminate SSL at this point. If you include a copy of Nginx as a reverse proxy, as is standard
in kubernetes, nginx can terminate SSL for you. I'll get to SSL, just be patient with me ❤️

# Feature Specifics
## Andrew's Custom Page Elements
### Valid Go Template Instructions for Rendering Page Structures
```text
.AndrewTableOfContents
.AndrewTableOfContentsWithDirectories
```
These are for generating lists of web pages that exist at the same level in the file system as the web page and in child directories. 

Andrew sorts by page publish date. This publish date is tricky for a file-based web server to get consistent, so here's the rules:
1. If you have the tag `<meta name="andrew-publish-time" content="YYYY-MM-DD"/>`, Andrew uses this date.
2. Andrew uses the page's mtime. This means that if you edit a page that does not contain the `andrew-publish-time` element, then you will push it back to the top of the list.

If your page contains an `andrew-publish-time` meta element, the time must be formatted in YYYY-MM-DD format. Minutes and hours aren't supported yet.
I don't write a lot, so I don't need granularity beyond a single day. Adding finer granularity isn't hard; feel free to ask for it or write a PR.


### Semantically Meaningful Andrew-specific HTML elements
```html    
<meta name="andrew-publish-time" content="YYYY-MM-DD"/>
<title>Your page title</title>
```

All `meta` elements are actually parsed in the [Andrew Page](./page.go), but Andrew doesn't use a lot of them just yet.

### Custom CSS IDs and classes
I've tried to consistently include the string `andrew` in front of any CSS classes or IDs, so they're less likely to
clash with your whimsy for laying out your own site.

The reason these classes and IDs exist is simple: it makes it easier for you to style Andrew's unstyled HTML. I don't want
Andrew making decisions about your website's layout.

I include classes and IDs that get my sites looking how I want. If you need more, file a request.

### How does the .AndrewTableOfContents render?
AndrewTableOfContents is for rendering a table of contents of the pages beneath the current page. It only lists page links.
If you want your links grouped by directories, check out `.AndrewTableOfContentsWithDirectories`.

This is handled in linksbuilder.go. I try to keep this README up to date, but if seems like it doesn't sync with reality
the final word is the source code.

Given this file system structure:
```text
index.html
articles/
        index.html
        article-1.html
        article-2.html
        article-2.css
        article-1.js
fanfics/
        index.html
        story-1/
                potter-and-draco.html
        story-2/
                what-if-elves-rode-mice-pt1.html
                what-if-elves-rode-mice-pt2.html
```

if articles/index.html contains `{{ .AndrewTableOfContents }}` anywhere, that will be replaced with:

```html
    <a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="article-1.html">article 1</a>  - <span class=\"publish-date\">0000-00-01</span></li>
    <a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="article-2.html">article 2</a>  - <span class=\"publish-date\">0000-00-01</span></li>
```

if fanfics/index.html contains `{{ .AndrewTableOfContents }}`, that'll be replaced with:

```html
    <a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="story-1/potter-and-draco.html">Potter and Draco</a>  - <span class=\"publish-date\">0000-00-01</span></li>
    <a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt1.html</a>  - <span class=\"andrew-page-publish-date\">0000-00-01</span></li>
    <a class="andrewtableofcontentslink" id="andrewtableofcontentslink2" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt2.html</a>  - <span class=\"andrew-page-publish-date\">0000-00-01</span></li>
```


## how is the .AndrewTableOfContentsWithDirectories rendered?
Given this file system structure:
```text
groupedContents.html
articles/
        index.html #this will be ignored. index.html normally contains its own listing of pages, but this is already a page list.
        article-1.html
        article-2.css #this will be ignored; Andrew only links to html files.
        articles-series/
                dragons-are-lovely.html
                dragons-are-fierce.html
```
if index.html contains `{{ .AndrewTableOfContentsWithDirectories }}` anywhere, that will be replaced with a `<div>` called AndrewTableOfContentsWithDirectories.
Inside the `<div>` is a decent representation of all of your content:

```html
<div class="AndrewTableOfContentsWithDirectories">
<ul>
        <h5>articles/</h5>
        <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="articles/article-1.html">article-1.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>

<ul>
        <h5><span class="AndrewParentDir">articles/</span>articles-series/</h5>
        <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="articles/articles-series/dragons-are-lovely.html">dragons-are-lovely.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
        <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="articles/articles-series/dragons-are-fierce.html">dragons-are-fierce.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>

</div>
```

Note the inclusion of a `<span>` around the name of the parent directory. The parent directory name is a bit repetitive, so I wanted to
be able to style it to not draw attention to it.

If the above seems out of sync with reality, the easiest place to get a canonical representation of what Andrew's building will be
in [linksbuilder_test.go](./linksbuilder_test.go)

## page titles
If a page contains a `<title>` element, Andrew picks it up and uses that as the name of a link.
If the page does not contain a `<title>` element, then Andrew will use the file name of that file as the link name.

## meta elements
Andrew parses meta tags and makes them accessible on its AndrewPage object.

### valid meta elements
<meta name="andrew-publish-time" value="2024-03-12">

## sitemap.xml
When the endpoint `baseUrl/sitemap.xml` is visited, Andrew will automatically generate a sitemap containing paths to all html pages.

