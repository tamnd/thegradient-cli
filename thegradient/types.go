package thegradient

// Post is one article from The Gradient.
type Post struct {
	Rank   int    `json:"rank"   csv:"rank"   tsv:"rank"`
	Date   string `json:"date"   csv:"date"   tsv:"date"`
	Author string `json:"author" csv:"author" tsv:"author"`
	Title  string `json:"title"  csv:"title"  tsv:"title"`
	URL    string `json:"url"    csv:"url"    tsv:"url"`
}

// Info holds aggregate statistics for The Gradient.
type Info struct {
	TotalPosts int    `json:"total_posts"`
	OldestPost string `json:"oldest_post"`
	LatestPost string `json:"latest_post"`
	FeedURL    string `json:"feed_url"`
	SiteURL    string `json:"site_url"`
}
