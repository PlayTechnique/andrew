# andrew

A web server, like web servers used to be.

Andrew is a web server for the Indie Web. It has the convenience of a static site generator without a build step; it has the niceties of a CMS like an RSS feed, a sitemap, a self-maintaining table of contents, with no databases attached.

You write your html, css and javascript on your file system, Andrew shows them to the world. Andrew has no opinions about frameworks, libraries, interactivity. Andrew's just a web server.

In 1991, Tim Berners-Lee wrote an HTML file and put it on a file system. 30 years later, linking two pages together needs a database and a build pipeline and a frontend framework. Andrew is what you'd build if you had to explain that evolution to Ken Thompson and felt embarrassed.

## Demo

No database. No build step. No framework. No opinions about your CSS. Just some lovely html, and a little template syntax.

```bash
; mkdir mysite && cd mysite
; echo '<body><p>Check out these great child pages:</p>{{ .AndrewTableOfContents }}' > index.html
; echo "<body>childpage</body>" > no-bloat.html
; echo "<body>childpage</body>" > even-less-bloat.html
; andrew . 127.0.0.1:8080 localhost &
[1] 99601
; curl 127.0.0.1:8080

<body><p>Check out these great child pages:</p>
<div class="AndrewTableOfContents"> # <----- This div was written by Andrew.
<ul>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="even-less-bloat.html">even-less-bloat.html</a> - <span class="andrew-page-publish-date">2026-05-17</span></li>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="no-bloat.html">no-bloat.html</a> - <span class="andrew-page-publish-date">2026-05-17</span></li>
</ul>
</div>
```

- You get started by writing an "index.html" file and then running Andrew from that directory.
- You can optionally provide an ssl certificate and key. No need for nginx or something to front Andrew.
- Convenient support for creating lists of files from the current directory down, so you don't need to maintain that list by hand. You control
  the specific order of the pages in the list by simply providing an html meta tag in your `<head>` element
- Sitemap automatically generated from your file system layout, as search engines like to see this for brand new websites.
- RSS Feed automatically generated from your file system layout. Folks can subscript at `baseUrl/rss.xml`

I wanted an http server that allows me to add a simple go template instruction into an index.html that is replaced
with the contents of any html files that are below the current index.html in the file system hierarchy.

Andrew contains the concept of an AndrewPage. This structure makes various pieces of metadata stored within your web page
available to Andrew for creating links and sorting pages in the various tables of contents available (see below). The specifics
are explained below, but conceptually I'm trying to use standard html elements to inform Andrew about site metadata. For more
you may want to check the [Architecture.md](./ARCHITECTURE.md)

## To run it

`go run github.com/playtechnique/andrew/cmd/andrew@latest` is the simple way. The available versions are all git tags.

There are github releases, too. You can find compiled binaries at https://github.com/PlayTechnique/andrew/releases for
linux, windows, and macOS for both amd64 and arm64 on all systems.

Here's how I install andrew through Docker:

```
FROM golang:1.23 AS base

WORKDIR /usr/src/app

ENV CGO_ENABLED=0
RUN go install github.com/playtechnique/andrew/cmd/andrew@latest

FROM scratch

COPY --from=base /etc/passwd /etc/passwd
USER 1000
COPY --from=base /go/bin/andrew /andrew
COPY --chown=1000:1000 content /website

EXPOSE 8080

ENTRYPOINT ["/andrew"]
```

## Arguments and Options

arguments are mandatory. Options aren't.

### Arguments

andrew accepts up to three arguments, in this order:

```bash
andrew [contentRoot] [address] [baseUrl]
```

contentRoot is the directory you're serving from, that contains your top level index.html. andrew follows
apache's lead on expecting index.html in any directory as a default page.

address is the address you want to bind the server to. Specify as an address:port combination.

baseUrl is the hostname you're serving from. This is a part of sitemaps and future rss feeds. It also contains the protocol
e.g. `https://playtechnique.io`

### Options

-h | --help - show the help

-c | --cert - this is a paired option with the option below. The path to an SSL cert bundle.

-p |--privatekey - this is paired with the option above. The path to your ssl private key.

-d |--rssdescription - a short description of your RSS feed.

-t |--rsstitle - the title for your RSS feed.

# Feature Specifics

## SSL Support

Want to serve your site over https? So does everyone else!
Start up andrew with the arguments `--cert` and `--privatekey`.
If you forget one of them, but supply the other, you'll get a helpful error reminding you what you need to do.
Andrew happily serves over https. It also serves over http.

## Andrew's Custom Page Elements

### Valid Go Template Instructions for Rendering Page Structures

```text
.AndrewTableOfContents
.AndrewTableOfContentsWithDirectories
```

These are for generating lists of web pages that exist at the same level in the file system as the web page and in child directories.

Note that each of these creates its items inside a div. Here's your cheat sheet:

```text
.AndrewTableOfContents becomes a div with class AndrewTableOfContents
.AndrewTableOfContentsWithDirectories becomes a div with class AndrewTableOfContentsWithDirectories
```

Andrew sorts by page publish date. This publish date is tricky for a file-based web server to get consistent, so here's the rules:

1. If you have the tag `<meta name="andrew-publish-time" content="YYYY-MM-DD"/>`, Andrew uses this date e.g. <meta name="andrew-publish-time" content="2025-03-30"/>

This date format can be generated with the cli command `date +%Y-%m-%d`, so you can pretty easily use a shell alias to generate this on the fly for a new page. 2. If you have the tag `<meta name="andrew-publish-time" content="YYYY-MM-DD HH:MM:SS"/>`, Andrew refines the date with the time published. This allows you to publish several articles on the same day and get the ordering correct e.g. `<meta name="andrew-publish-time" content="2025-03-30 12:30:00">`.

This date format can be generated with the cli command `date +"%Y-%m-%d %H:%M:%S`. 3. If the `andrew-publish-time` meta tag is not present, Andrew uses the page's mtime. This means that if you edit a page that does not contain the `andrew-publish-time` element, then you will push it back to the top of the table of contents. This is the worst solution if you're using andrew in a container and building the web pages in as part of your build pipeline. I'm doing this, so I use `andrew-publish-time` a lot.

If you want to automate generating the datestamp with timestamp, this'll get you where you want to be on macOS or linux `date +"%Y-%m-%d %H:%M:%S"`

### Semantically Meaningful Andrew-specific HTML elements

```html
<meta name="andrew-publish-time" content="YYYY-MM-DD" /> <title>Your page title</title>
```

All `meta` elements are actually parsed in the [Andrew Page](./page.go), but Andrew doesn't use a lot of them just yet.

### Custom CSS IDs and classes

When the table of contents is output to your page, it's a block of html that needs styling. The layout is drawn in [renderAndrewTableOfContentsWithDirectories](./page.go#L123) and [renderAndrewTableOfContents](https://github.com/PlayTechnique/andrew/blob/156b6db7de9e22e32b0f57f92599f558748706bc/linksbuilder.go#L176), but the layout and structure is detailed in an example below.

I've tried to consistently include the string `andrew` in front of any CSS classes or IDs, so they're less likely to clash with your own class names. Again, example below of class and ID names are available for styling. I don't want Andrew making decisions about your website's layout.

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

if articles/index.html contains `{{ .AndrewTableOfContents }}` anywhere, that will be replaced with a div like this one:

```html
    <div class="AndrewTableOfContents">
    <ul>
    <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="article-1.html">article 1</a>  - <span class=\"publish-date\">0000-00-01</span></li>
    <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="article-2.html">article 2</a>  - <span class=\"publish-date\">0000-00-01</span></li>
    </ul>
    </div>
```

if fanfics/index.html contains `{{ .AndrewTableOfContents }}`, that'll be replaced with a div like this one:

```html
    <div class="AndrewTableOfContents">
    <ul>
    <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="story-1/potter-and-draco.html">Potter and Draco</a>  - <span class=\"publish-date\">0000-00-01</span></li>
    <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt1.html</a>  - <span class=\"andrew-page-publish-date\">0000-00-01</span></li>
    <li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink2" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt2.html</a>  - <span class=\"andrew-page-publish-date\">0000-00-01</span></li>
    </ul>
    </div>
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
Inside the `<div>` is a decent representation of all of your content. The order of directories is determined by the most recent content in each directory, so
the directory with the most recent content will be the first one in the list:

```html
<div class="AndrewTableOfContentsWithDirectories">
  <ul>
    <h5>articles/</h5>
    <li>
      <a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="articles/article-1.html"
        >article-1.html</a
      >
      - <span class="andrew-page-publish-date">0001-01-01</span>
    </li>
  </ul>

  <ul>
    <h5><span class="AndrewParentDir">articles/</span>articles-series/</h5>
    <li>
      <a
        class="andrewtableofcontentslink"
        id="andrewtableofcontentslink1"
        href="articles/articles-series/dragons-are-lovely.html"
        >dragons-are-lovely.html</a
      >
      - <span class="andrew-page-publish-date">0001-01-01</span>
    </li>
    <li>
      <a
        class="andrewtableofcontentslink"
        id="andrewtableofcontentslink1"
        href="articles/articles-series/dragons-are-fierce.html"
        >dragons-are-fierce.html</a
      >
      - <span class="andrew-page-publish-date">0001-01-01</span>
    </li>
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

## rss.xml

When the endpoint `baseUrl/rss.xml` is visited, Andrew will automatically generate an RSS feed with all your articles in! We love an RSS feed <3
