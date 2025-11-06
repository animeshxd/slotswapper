package api

import (
	"net/http"
	"path"
)

func (s *Server) HandleReactFiles(staticDir string) http.HandlerFunc {
	root := http.Dir(staticDir)
	fs := http.FileServer(root)
	indexPath := path.Join(staticDir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		f, err := root.Open(path.Clean(r.URL.Path))
		if err != nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil || info.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}

		fs.ServeHTTP(w, r)
	}
}
