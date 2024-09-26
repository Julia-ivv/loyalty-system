package middleware

import (
	"net/http"
	"strings"

	"github.com/Julia-ivv/loyalty-system.git/internal/app/compressing"
)

func HandlerWithGzipCompression(h http.Handler) http.Handler {
	contentTypeForCompression := [2]string{"application/json", "text/html"}
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			newRes := res
			acceptEncoding := req.Header.Values("Accept-Encoding")
			supportGzip := false
			for _, v := range acceptEncoding {
				if strings.Contains(v, "gzip") {
					supportGzip = true
				}
			}
			contentType := req.Header.Get("Content-Type")
			needsCompressing := false
			for _, v := range contentTypeForCompression {
				if strings.Contains(contentType, v) {
					needsCompressing = true
					break
				}
			}

			if supportGzip && needsCompressing {
				zw := compressing.NewGzipWriter(res)
				newRes = zw
				defer zw.Writer.Close()
				zw.Header().Set("Content-Encoding", "gzip")
			}

			contentEncoding := req.Header.Get("Content-Encoding")
			gzipEncoding := strings.Contains(contentEncoding, "gzip")
			if gzipEncoding {
				zr, err := compressing.NewGzipReader(req.Body)
				if err != nil {
					http.Error(newRes, err.Error(), http.StatusInternalServerError)
					return
				}
				req.Body = zr
				defer zr.Close()
			}

			h.ServeHTTP(newRes, req)
		})
}
