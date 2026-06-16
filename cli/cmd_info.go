package cli

import "github.com/spf13/cobra"

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show The Gradient publication statistics",
		Long: `info prints aggregate statistics: total articles, oldest and latest post
dates, and the feed URL.

Examples:
  gradient info
  gradient info -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info, err := a.client.Stats(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
}
