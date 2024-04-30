# andrew

I wanted an http server that allows me to add a simple annotation into an index.html that is replaced
with the contents of any html files that are below the current index.html in the file system hierarchy.

It's grown a little to include a small sitemap generator.  

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

baseUrl is the hostname you're serving from. This is a part of sitemaps and rss feeds. It contains the protocol
e.g. `https://playtechnique.io`


## rendering the .AndrewIndexBody
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

if articles/index.html contains `{{ .AndrewIndexBody }}` anywhere, that will be replaced with:

```html
    <a class="andrewindexbodylink" id="andrewindexbodylink0" href="article-1.html">article 1</a>
    <a class="andrewindexbodylink" id="andrewindexbodylink1" href="article-2.html">article 2</a>
```

if fanfics/index.html contains `{{ .AndrewIndexBody }}`, that'll be replaced with:

```html
    <a class="andrewindexbodylink" id="andrewindexbodylink0" href="story-1/potter-and-draco.html">Potter and Draco</a>
    <a class="andrewindexbodylink" id="andrewindexbodylink1" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt1.html</a>
    <a class="andrewindexbodylink" id="andrewindexbodylink2" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt2.html</a>
```

## page titles
If a page contains a `<title>` element, Andrew picks it up and uses that as the name of a link.
If the page does not contain a `<title>` element, then Andrew will use the file name of that file as the link name.

## meta elements
Andrew parses meta tags and makes them accessible on its AndrewPage object.

For a meta element to be picked up, it must be formatted with andrew- prepending the meta element's name, like this `<meta name="andrew-<rest of the name>" value="your-value">`

### valid meta elements
<meta name="andrew-created-on" value="2024-03-12">
<meta name="andrew-tag" value="diary entry">

## ordering of pages
If a page contains the meta element `<meta name=andrew-created-on value="2024-03-12">` in its `<head>`, Andrew orders on these tags.
If the page does not contain the meta element, it uses the mtime of the file to try and determine ordering. This means that if you edit a page
that does not contain the `andrew-created-on` element, then you will push it back to the top of the list.

If your page contains an `andrew-created-on` meta element, the time must be formatted in accordance with <SOME STANDARD HERE>. If your `andrew-created-on`
contains a date but not a time, Andrew assumes the page was created at midnight.

## sitemap.xml
When the endpoint `baseUrl/sitemap.xml` is visited, Andrew will automatically generate a sitemap containing paths to all html pages.

