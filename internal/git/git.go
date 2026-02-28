package git

// Client wraps git CLI operations.
type Client struct {
	repoPath string
}

// NewClient creates a new git client for the given repository path.
func NewClient(repoPath string) *Client {
	return &Client{repoPath: repoPath}
}
