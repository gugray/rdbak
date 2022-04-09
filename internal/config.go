package internal

// // {"email": "gabor.reads.this@gmail.com", "password": "aM!<Yk`ohQ_^=5U\",<vt"}

type Config struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	BookmarksFile string `json:"bookmarksFile"`
	ExportDir     string `json:"exportDir"`
}
