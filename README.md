tempshare
=====

A very simple tool to share files over HTTP for a limited number of times. Files are servd directly, and folders are zipped on the fly before being served.

Generates a random URL to serve the content, and stops when the file/folder has been served the specified number of times.

Serves a HTTP 401 Unauthorized for any request not on the generated URL. Pairs well with [blockfast](https://github.com/pldubouilh/blockfast).

```txt
Tempshare - CLI tool to share a file or folder a limited number of times

An unique URL will be generated, and the file/folder will be served at this path
The file/folder will be served a limited number of times, then the server will stop
Files are served directly, folder are zipped on the fly before being served

Usage: tempshare [options] <file/folder>
  -h string
        host to listen to (default "127.0.0.1")
  -p string
        port to listen to (default "8005")
  -s int
        will be shared n times (default 2)
```