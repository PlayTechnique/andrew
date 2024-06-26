* ~~how do I handle images?~~
* ~~how do I handle css? The mime type is being inconsistently set, so when I am going to serve a css file the mime type is text/plain ~~
    ~~but sometimes the web browser gives me the benefit of the doubt.~~
* ~~How do I handle anything that's not html, for that matter? - the relative links might be right.~~
* ~~switch the test paradigm from the funky go routine + http server to use the test library~~
* read and order first on andrew-published-at and then on date. Does this work okay in a container?
* Documentation of functionality
~~* select random free port in tests~~
* generate sitemap
* there's both a page parser and a page formatter. They shouldn't be spread through other functions, but should be isolated into a file.
* github workflow for homebrew
* html escape the paths you're serving
* pull out any article summaries into parent card
* extract the function that builds the index body out of serveIndexPage and pass it in instead. You can also pass in AndrewTableOfContents; that
    would give you the flexibility to render different kinds of functions based upon the presence of different template strings.