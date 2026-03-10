package cmdutil

// PagedResponse is used for agent mode paged responses.
type PagedResponse struct {
	Items      any    `json:"items"`
	NextCursor string `json:"next_cursor"`
}
