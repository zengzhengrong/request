/*
Copyright © 2023 Zengzhengrong <bhg889@163.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zengzhengrong/request/opts/client"
	"github.com/zengzhengrong/request/response"
)

const (
	GET    string = "GET"
	POST   string = "POST"
	PUT    string = "PUT"
	PATCH  string = "PATCH"
	DELETE string = "DELETE"
)

// matchCondition is check is match expect
func matchCondition(resp response.Response, expectStatusCode int, expectType, expectedPath string) bool {
	switch expectType {
	case "header":
	case "json":
	default:
		// default is check statuscode
		if resp.Resp.StatusCode == expectStatusCode {
			return true
		}
	}
	return false
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zurl",
	Short: "zurl is similar curl but have more function",
	Long:  `zurl is similar curl but have more function`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			expectType   string = ""
			expectedPath string = ""
		)
		method, err := cmd.Flags().GetString("method")
		if err != nil {
			return err
		}
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}
		retry, err := cmd.Flags().GetInt("retry")
		if err != nil {
			return err
		}
		interval, err := cmd.Flags().GetInt("interval")
		if err != nil {
			return err
		}
		header, err := cmd.Flags().GetStringToString("add-header")
		if err != nil {
			return err
		}
		query, err := cmd.Flags().GetStringToString("add-query")
		if err != nil {
			return err
		}
		expectHeader, err := cmd.Flags().GetString("expect-header")
		if err != nil {
			return err
		}
		expectJson, err := cmd.Flags().GetString("expect-json")
		if err != nil {
			return err
		}
		expectStatusCode, err := cmd.Flags().GetInt("expect-statuscode")
		if err != nil {
			return err
		}
		if expectHeader == "" && expectJson == "" && expectStatusCode == 0 {
			return fmt.Errorf(`"expect-header" or "expect-json" and "expect-statuscode", one of them must be specified`)
		}
		if expectHeader != "" && expectJson != "" {
			return fmt.Errorf(`can not both set "expect-header" and "expect-json"`)
		}
		if expectHeader != "" {
			expectType = "header"
			expectedPath = expectHeader
		} else if expectJson != "" {
			expectType = "json"
			expectedPath = expectHeader
		}

		client := client.NewClient(
			client.WithTimeOut(3600 * time.Second),
		)
	loop:
		for i := 0; i < retry; i++ {
			switch strings.ToUpper(method) {
			case GET:
				resp := client.GET(url, query, header)
				if matchCondition(resp, expectStatusCode, expectType, expectedPath) {
					fmt.Println("Success match condition , exit ...")
					break loop
				}
			case POST:
			case PUT:
			case PATCH:
			case DELETE:
			default:
				return fmt.Errorf("%s method does not math [GET,POST,PUT,PATCH,DELETE] ", method)
			}
			fmt.Printf("Failed match condition , retry %d ...", i)
		}
		fmt.Println(method, url, retry, expectHeader, expectJson, interval, header, query, args)
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of zurl",
	Long:  `All software has versions`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version v1.0")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().StringP("method", "m", "GET", "--method specify Http Method")
	rootCmd.Flags().String("url", "", "--url specify url")
	rootCmd.Flags().IntP("retry", "r", 0, "--retry retry count , 0 is never stop")
	rootCmd.Flags().StringToString("add-header", map[string]string{}, "--add-header , add the header to request")
	rootCmd.Flags().StringToString("add-query", map[string]string{}, "--add-query , add the query args to request")
	rootCmd.Flags().IntP("expect-statuscode", "s", 200, "--expect-statuscode get expected statuscode result ,retry if not equal ")
	rootCmd.Flags().String("expect-header", "", "--expect-header get expected header result ,retry if not equal ")
	rootCmd.Flags().String("expect-json", "", "--expect-header get expected json result ,use xx.xx to specify value,retry if not equal ")
	rootCmd.Flags().IntP("interval", "i", 1, "--interval ,The interval between retries")

}
