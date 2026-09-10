package payagentic

// ListOptions configures paginated list requests.
type ListOptions struct {
	// Cursor is an opaque pagination cursor for fetching the next page.
	Cursor string `json:"cursor,omitempty"`
	// Limit is the maximum number of items to return per page.
	Limit int `json:"limit,omitempty"`
	// Filter is an optional filter expression.
	Filter string `json:"filter,omitempty"`
}

// queryParams returns URL query parameters for the list options.
func (o *ListOptions) queryParams() string {
	if o == nil {
		return ""
	}
	params := ""
	sep := "?"
	if o.Cursor != "" {
		params += sep + "cursor=" + o.Cursor
		sep = "&"
	}
	if o.Limit > 0 {
		params += sep + "limit=" + itoa(o.Limit)
		sep = "&"
	}
	if o.Filter != "" {
		params += sep + "filter=" + o.Filter
	}
	return params
}

// itoa converts an int to a string without importing strconv in the hot path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// Reverse.
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

// Iterator provides a convenient way to iterate over all pages of a paginated response.
type Iterator[T any] struct {
	fetch  func(cursor string) (*PaginatedResponse[T], error)
	cursor string
	done   bool
}

// NewIterator creates a new paginated iterator. The fetch function is called for each page
// with the current cursor value (empty string for the first page).
func NewIterator[T any](fetch func(cursor string) (*PaginatedResponse[T], error)) *Iterator[T] {
	return &Iterator[T]{fetch: fetch}
}

// Next returns the next page of results. It returns nil, nil when all pages have been consumed.
func (it *Iterator[T]) Next() (*PaginatedResponse[T], error) {
	if it.done {
		return nil, nil
	}
	resp, err := it.fetch(it.cursor)
	if err != nil {
		return nil, err
	}
	if resp.NextCursor == "" {
		it.done = true
	} else {
		it.cursor = resp.NextCursor
	}
	return resp, nil
}
