package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func (a *App) exportCmd() *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all articles from The Gradient as JSONL",
		Long: `export fetches all articles from The Gradient RSS feed and writes one JSON
record per line. Use --out to write to a file instead of stdout.

Examples:
  gradient export > gradient.jsonl
  gradient export --out gradient.jsonl
  gradient export -o csv > gradient.csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			posts, err := a.client.AllPosts(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return codeError(exitError, err)
				}
				defer f.Close()
				r := a.newRendererTo(f)
				return r.Render(posts)
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	cmd.Flags().StringVar(&outFile, "out", "", "write output to FILE instead of stdout")
	return cmd
}
