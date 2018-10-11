// Package seektar dynamically generates tarballs in a
// deterministic way that allows seeking.
// This can be used on servers to provide downloads for an
// entire directory, while still supporting Byte-Range
// HTTP requests.
//
// Example:
//
//     tarResult, _ := seektar.Tar("/path/to/directory", "directory")
//     tarFile, _ := tarResult.Open()
//     defer tarFile.Close()
//     // tarFile is an io.Reader, io.Seeker, and io.Closer.
//     // It dynamically generates tar data.
//
package seektar
