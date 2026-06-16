// Package thegradient is the library behind the thegradient CLI.
package thegradient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const DefaultUserAgent = "thegradient-cli/dev (+https://github.com/tamnd/thegradient-cli)"

type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://thegradient.pub",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var (
	itemRe    = regexp.MustCompile(`(?s)<item>(.*?)</item>`)
	cdataRe   = regexp.MustCompile(`(?s)<!\[CDATA\[(.*?)\]\]>`)
	titleRe   = regexp.MustCompile(`(?s)<title>(.*?)</title>`)
	linkRe    = regexp.MustCompile(`<link>([^<]+)</link>`)
	dateRe    = regexp.MustCompile(`<pubDate>([^<]+)</pubDate>`)
	authorRe  = regexp.MustCompile(`(?s)<(?:dc:creator|author)>(.*?)</(?:dc:creator|author)>`)
	tagRe     = regexp.MustCompile(`<[^>]+>`)
)

func extractText(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func extractCDATA(s string) string {
	if m := cdataRe.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(s)
}

// Top fetches the RSS feed page and returns posts.
func (c *Client) Top(ctx context.Context, page int) ([]*Post, error) {
	url := c.cfg.BaseURL + "/rss/"
	if page > 1 {
		url = fmt.Sprintf("%s/rss/?page=%d", c.cfg.BaseURL, page)
	}
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	xml := string(body)

	var posts []*Post
	rank := (page-1)*15 + 1
	for _, m := range itemRe.FindAllStringSubmatch(xml, -1) {
		item := m[1]

		tm := titleRe.FindStringSubmatch(item)
		if tm == nil {
			continue
		}
		lm := linkRe.FindStringSubmatch(item)
		dm := dateRe.FindStringSubmatch(item)
		am := authorRe.FindStringSubmatch(item)

		title := extractCDATA(tm[1])
		link := ""
		if lm != nil {
			link = strings.TrimSpace(lm[1])
		}
		date := ""
		if dm != nil {
			date = strings.TrimSpace(dm[1])
			// Normalize date: "Wed, 18 Feb 2026 23:25:52 GMT" -> "2026-02-18"
			if t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", date); err == nil {
				date = t.Format("2006-01-02")
			}
		}
		author := ""
		if am != nil {
			author = extractText(extractCDATA(am[1]))
		}

		posts = append(posts, &Post{
			Rank:   rank,
			Title:  title,
			Author: author,
			Date:   date,
			URL:    link,
		})
		rank++
	}
	return posts, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// AllPosts fetches the first RSS page and returns all posts.
func (c *Client) AllPosts(ctx context.Context) ([]*Post, error) {
	return c.Top(ctx, 1)
}

// Stats returns aggregate statistics from the RSS feed.
func (c *Client) Stats(ctx context.Context) (*Info, error) {
	posts, err := c.Top(ctx, 1)
	if err != nil {
		return nil, err
	}
	info := &Info{
		TotalPosts: len(posts),
		FeedURL:    c.cfg.BaseURL + "/rss/",
		SiteURL:    c.cfg.BaseURL,
	}
	if len(posts) > 0 {
		info.LatestPost = posts[0].Date
		info.OldestPost = posts[len(posts)-1].Date
	}
	return info, nil
}
