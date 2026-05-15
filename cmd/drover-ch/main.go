package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud-shuttle/drover-ch-optimiser/internal/analyzer"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/chclient"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/config"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/graph"
	"github.com/cloud-shuttle/drover-ch-optimiser/internal/report"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	verbose    bool
	outputFmt  string
	outputFile string
	focusTable string
	database   string
	daysBack   int

	chHost     string
	chPort     int
	chUser     string
	chPassword string
	chSecure bool
)

var rootCmd = &cobra.Command{
	Use:   "drover-ch",
	Short: "ClickHouse warehouse auditor and optimizer",
	Long: `drover-ch-optimiser audits ClickHouse warehouses against official
best practices, surfaces optimization opportunities, and visualizes
dependency and query-usage graphs.

Part of the Drover family — static, read-only, agent-ready.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "rules" {
			return nil
		}
		if err := config.Load(cfgFile); err != nil {
			return err
		}
		var secure *bool
		if cmd.Root().PersistentFlags().Changed("secure") {
			secure = &chSecure
		}
		config.ApplyOverrides(config.Overrides{
			Host:     chHost,
			Port:     chPort,
			Username: chUser,
			Password: chPassword,
			Secure:   secure,
		})
		return nil
	},
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run full warehouse audit",
	RunE:  runAudit,
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate dependency graph",
	RunE:  runGraph,
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List best-practice rules",
	Run: func(cmd *cobra.Command, args []string) {
		analyzer.ListRules()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (optional; else ~/.drover/ch-optimiser.yaml if present)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.PersistentFlags().StringVar(&chHost, "host", "", "ClickHouse host (overrides config)")
	rootCmd.PersistentFlags().IntVar(&chPort, "port", 0, "ClickHouse native port (default 9000; Cloud often 9440 with --secure)")
	rootCmd.PersistentFlags().StringVar(&chUser, "user", "", "ClickHouse username (overrides config)")
	rootCmd.PersistentFlags().StringVar(&chPassword, "password", "", "ClickHouse password (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&chSecure, "secure", false, "use TLS (required for ClickHouse Cloud)")

	for _, c := range []*cobra.Command{auditCmd, graphCmd} {
		c.Flags().StringVarP(&outputFmt, "format", "f", "markdown", "output format: markdown, html, json")
		c.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")
		c.Flags().StringVar(&focusTable, "focus", "", "focus subgraph on db.table")
		c.Flags().IntVar(&daysBack, "days", 30, "days of query_log to analyze")
		c.Flags().StringVar(&database, "database", "", "limit scan to one database")
	}

	rootCmd.AddCommand(auditCmd, graphCmd, rulesCmd)

	suggestCmd := &cobra.Command{
		Use:   "suggest [db.table]",
		Short: "Optimization suggestions for one table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			focusTable = args[0]
			return runAudit(cmd, args)
		},
	}
	rootCmd.AddCommand(suggestCmd)
}

func graphOpts() graph.Options {
	return graph.Options{
		FocusTable: focusTable,
		DaysBack:   daysBack,
		Database:   database,
		Verbose:    verbose,
	}
}

func runAudit(cmd *cobra.Command, args []string) error {
	start := time.Now()

	conn, err := chclient.NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	if verbose {
		fmt.Println("Discovering tables and building graph...")
	}
	g, err := graph.BuildGraph(conn, graphOpts())
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("Running rules engine...")
	}
	issues, summary := analyzer.Run(g)
	rpt := report.New(g, issues, summary, daysBack)

	if outputFile == "" {
		switch outputFmt {
		case "html":
			outputFile = "clickhouse-audit.html"
		case "json":
			outputFile = "clickhouse-audit.json"
		default:
			outputFile = "clickhouse-audit.md"
		}
	}

	switch outputFmt {
	case "html":
		err = rpt.WriteHTML(outputFile)
	case "json":
		err = rpt.WriteJSON(outputFile)
	default:
		err = rpt.WriteMarkdown(outputFile)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Audit completed in %s\n", time.Since(start).Round(time.Millisecond))
	report.PrintSummary(summary, outputFile)
	return nil
}

func runGraph(cmd *cobra.Command, args []string) error {
	conn, err := chclient.NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	g, err := graph.BuildGraph(conn, graphOpts())
	if err != nil {
		return err
	}

	if outputFmt == "html" || outputFile != "" {
		if outputFile == "" {
			outputFile = "clickhouse-graph.html"
		}
		return report.WriteGraphOnlyHTML(g, outputFile)
	}

	fmt.Println(g.ToMermaid(graph.MermaidOpts{}))
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
