/*
Copyright © 2023 Zengzhengrong <bhg889@163.com>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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

type rootArgs struct {
	method           string
	url              string
	retry            int
	interval         int
	header           map[string]string
	query            map[string]string
	expectHeader     map[string]string
	expectJson       map[string]string
	expectStatusCode int
	expectType       string
	expectedData     map[string]string
	debug            bool
}

func newrootArgs(cmd *cobra.Command) (rootArgs, error) {
	var (
		expectType   string
		expectedData map[string]string
		debugp       bool
	)
	method, err := cmd.Flags().GetString("method")
	if err != nil {
		return rootArgs{}, err
	}
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return rootArgs{}, err
	}
	retry, err := cmd.Flags().GetInt("retry")
	if err != nil {
		return rootArgs{}, err
	}
	interval, err := cmd.Flags().GetInt("interval")
	if err != nil {
		return rootArgs{}, err
	}
	header, err := cmd.Flags().GetStringToString("add-header")
	if err != nil {
		return rootArgs{}, err
	}
	query, err := cmd.Flags().GetStringToString("add-query")
	if err != nil {
		return rootArgs{}, err
	}
	expectHeader, err := cmd.Flags().GetStringToString("expect-header")

	if err != nil {
		return rootArgs{}, err
	}
	expectJson, err := cmd.Flags().GetStringToString("expect-json")
	if err != nil {
		return rootArgs{}, err
	}
	expectStatusCode, err := cmd.Flags().GetInt("expect-statuscode")
	if err != nil {
		return rootArgs{}, err
	}
	debug, err := cmd.Flags().GetString("debug")
	if err != nil {
		return rootArgs{}, err
	}
	if debug != "" {
		debugp = true
	}

	if len(expectHeader) == 0 && len(expectJson) == 0 && expectStatusCode == 0 {
		return rootArgs{}, fmt.Errorf(`"expect-header" or "expect-json" and "expect-statuscode", one of them must be specified`)
	}
	if len(expectHeader) != 0 && len(expectJson) != 0 {
		return rootArgs{}, fmt.Errorf(`can not both set "expect-header" and "expect-json"`)
	}

	if len(expectHeader) != 0 {
		expectType = "header"
		expectedData = expectHeader
	} else if len(expectJson) != 0 {
		expectType = "json"
		expectedData = expectJson
	}

	return rootArgs{
		method:           method,
		url:              url,
		retry:            retry,
		interval:         interval,
		header:           header,
		query:            query,
		expectHeader:     expectHeader,
		expectJson:       expectJson,
		expectStatusCode: expectStatusCode,
		expectType:       expectType,
		expectedData:     expectedData,
		debug:            debugp,
	}, nil
}

// matchCondition is check is match expect
func matchCondition(resp response.Response, rootArgs rootArgs) bool {
	if resp.Err != nil {
		fmt.Printf("Faild to get response [err=%s]\n", resp.ErrorString())
		return false
	}
	switch rootArgs.expectType {
	case "header":
		n := len(rootArgs.expectHeader)
		if n == 0 {
			return false
		}
		mn := 0
		for key := range rootArgs.expectHeader {
			if rootArgs.expectHeader[key] == resp.Resp.Header.Get(key) {
				fmt.Printf("Math Header Success [key=%s][value=%s]\n", key, resp.Resp.Header.Get(key))
				mn++
			} else {
				fmt.Printf("Faild Math Header [key=%s][value=%s]\n", key, resp.Resp.Header.Get(key))
			}
		}
		if n == mn {
			return true
		}
	case "json":
		n := len(rootArgs.expectJson)
		if n == 0 {
			return false
		}
		mn := 0
		for key := range rootArgs.expectJson {
			value := gjson.GetBytes(resp.Body, key).String()
			if rootArgs.expectJson[key] == value {
				fmt.Printf("Math JsonBody Success [key=%s][value=%s]\n", key, value)
				mn++
			} else {
				fmt.Printf("Faild Math JsonBody [key=%s][value=%s]\n", key, value)
			}
		}
		if n == mn {
			return true
		}
	default:
		// default is check statuscode
		if resp.Resp.StatusCode == rootArgs.expectStatusCode {
			fmt.Printf("Math StatusCode Success %d,%d \n", rootArgs.expectStatusCode, resp.Resp.StatusCode)
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
	Run: func(cmd *cobra.Command, args []string) {
		var ismatch bool
		rootArgs, err := newrootArgs(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		client := client.NewClient(
			client.WithTimeOut(3600*time.Second),
			client.WithDebug(rootArgs.debug),
		)
	loop:
		for i := 0; i < rootArgs.retry; i++ {
			switch strings.ToUpper(rootArgs.method) {
			case GET:
				resp := client.GET(rootArgs.url, rootArgs.query, rootArgs.header)
				if matchCondition(resp, rootArgs) {
					fmt.Println("Success match condition , exit ...")
					ismatch = true
					break loop
				}
			case POST:
				err := errors.New("not support POST")
				fmt.Println(err)
				os.Exit(1)
			case PUT:
				err := errors.New("not support PUT")
				fmt.Println(err)
				os.Exit(1)
			case PATCH:
				err := errors.New("not support PATCH")
				fmt.Println(err)
				os.Exit(1)
			case DELETE:
				err := errors.New("not support DELETE")
				fmt.Println(err)
				os.Exit(1)
			default:
				err := fmt.Errorf("%s method does not math [GET,POST,PUT,PATCH,DELETE] ", rootArgs.method)
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Failed match condition , retry %d ...\n", i+1)
			time.Sleep(time.Duration(rootArgs.interval) * time.Second)
		}
		if ismatch {
			os.Exit(0)
		}
		err = errors.New("failed match condition and retry is max")
		fmt.Println(err)
		os.Exit(1)
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
	rootCmd.Flags().StringToString("expect-header", nil, "--expect-header get expected header result ,retry if not equal ")
	rootCmd.Flags().StringToString("expect-json", nil, "--expect-header get expected json result ,use xx.xx to specify value,retry if not equal ")
	rootCmd.Flags().IntP("interval", "i", 1, "--interval ,The interval between retries")
	rootCmd.Flags().StringP("debug", "d", "", "--debug ,Print more info for request or set env REQUEST_DEBUG=True")

}
