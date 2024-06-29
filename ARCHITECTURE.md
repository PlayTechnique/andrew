Andrew has two primary concepts: the Server and the Page.

The Server owns:
* answering questions about the layout of files in the directory structure
* understanding the kinds of file that are being served
* serving those files
If you are answering a question about files, the Server's got the answers. The server creates Pages
and serves those Pages.


The Page owns the content and metadata.

Page tracks the content of a specific file and various pieces of metadata about it.

For example, the page parses the contents of a file and parses out the andrew metadata headers, so that when the Server wants to present those elements to an end-user they're already built.