# seektar [![GoDoc](https://godoc.org/github.com/unixpickle/seektar?status.svg)](https://godoc.org/github.com/unixpickle/seektar)

Dynamically generated tarballs that support seeking. This can be used from a server to generate downloads for entire directories, while supporting Byte-Range requests.

# Example

Here's an example of how to use seektar:

```go
package main

import "github.com/unixpickle/seektar"

func main() {
    tarResult, _ := seektar.Tar("/path/to/directory", "directory")
    tarFile, _ := tarResult.Open()
    defer tarFile.Close()
    // tarFile is an io.Reader, io.Seeker, and io.Closer.
    // It dynamically generates tar data.
}
```
