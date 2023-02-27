package cmd

import (
	"fmt"

	"github.com/cnvergence/perf-goatz/pkg/k6"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var defaultK6flags = "-d 5m --rps 100 --system-tags=method,name,status,tag"

// execCmd represents the generate command
var execK6Cmd = &cobra.Command{
	Use:   "exec-k6",
	Short: "Execute K6 test suite",
	Long:  `Execute K6 test suite and download results`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		scriptPath, err := cmd.Flags().GetString("script")
		if err != nil {
			logrus.Errorf("Failed to get K6 script filename:%v", err)
			return err
		}
		k6Flags, err := cmd.Flags().GetString("flags")
		if err != nil {
			logrus.Errorf("Failed to get K6 flags:%v", err)
			return err
		}
		reportName, err := cmd.Flags().GetString("report")
		if err != nil {
			logrus.Errorf("Failed to get report filename:%v", err)
			return err
		}
		k6Cmd := fmt.Sprintf(`k6 run %s --out influxdb=http://test-influxdb:8086/k6 %s`, scriptPath, k6Flags)
		cmds := []string{
			"sh",
			"-c",
			k6Cmd,
		}
		config := k6.NewConfig(ctx)

		logrus.Infof("Executing command: %s", k6Cmd)
		stdout, stderr, err := config.Exec(ctx, cmds)
		if err != nil {
			logrus.Errorf("Failed to execute command:%v", err)
			return err
		}

		logrus.Printf("%s", stdout)
		logrus.Printf("%s", stderr)

		err = config.DownloadReport(reportName)
		if err != nil {
			logrus.Errorf("Failed to download report:%v", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execK6Cmd)
	rootCmd.PersistentFlags().StringP("script", "s", "k6_script.js", "Path to the K6 .js script file")
	rootCmd.PersistentFlags().StringP("report", "r", "summary.html", "K6 test HTML report filename")
	rootCmd.PersistentFlags().StringP("flags", "f", defaultK6flags, "K6 default flags")
}
