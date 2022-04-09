package internal

type Config struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"encryptedPassword"`
	BookmarksFile     string `json:"bookmarksFile"`
	ExportDir         string `json:"exportDir"`
}
