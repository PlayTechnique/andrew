# andrew

I wanted an http server that allows me to add a simple annotation into an index.html that is replaced
with the contents of any html files that are below the current index.html in the file system hierarchy.

It's grown a little to include a small sitemap generator.  

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
    <a class="andrewindexbodylink" id="andrewindexbodylink0" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt1.html</a>
    <a class="andrewindexbodylink" id="andrewindexbodylink0" href="story-2/what-if-elves-rode-mice-pt1.html">what-if-elves-rode-mice-pt2.html</a>
```

## page titles
If a page contains a `<title>` element, Andrew picks it up and uses that as the name of a link.
If the page does not contain a `<title>` element, then Andrew will use the file name of that file as the link name.

## ordering of pages
In this release, Andrew orders your page links asci-betically.

## server
The quickest way to get up and running is to cd into a directory containing web pages and `go run github.com/playtechnique/andrew/cmd/andrew@v0.0.4 .`. It binds to port 8080.
