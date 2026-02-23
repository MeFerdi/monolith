package document
import(
	"time"
)

type Document struct {
	ID           string
	Title        string
	Source       string // "local", "gdrive", "gdocs", "notion"
	Path         string //filepath or URL
	Content      string
	ContentHash  string
	LastModified time.Time
	MimeType     string
}
