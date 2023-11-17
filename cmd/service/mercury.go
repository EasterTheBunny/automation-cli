package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	mercuryCmd.Flags().StringVarP(&listen, "listen", "l", "127.0.0.1", "listen address")
	mercuryCmd.Flags().Int64VarP(&port, "port", "p", 8080, "listen port")
}

var (
	listen string
	port   int64

	mercuryCmd = &cobra.Command{
		Use:   "mercury",
		Short: "Run a mocked mercury server",
		Long:  `Run a mocked mercury server`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var handler http.HandlerFunc = func(writer http.ResponseWriter, reader *http.Request) {
				_ = reader.ParseForm()

				fmt.Fprintf(cmd.OutOrStdout(), "MercuryHTTPServe:RequestURI: %s", reader.RequestURI)

				for key, value := range reader.Form {
					fmt.Fprintf(cmd.OutOrStdout(), "MercuryHTTPServe:FormValue: key: %s; value: %s;", key, value)
				}

				report := MercuryV02Response{
					ChainlinkBlob: DefaultMercuryV2Report,
				}

				bts, err := json.Marshal(report)
				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)

					return
				}

				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write(bts)
			}

			server := &http.Server{
				Addr:              fmt.Sprintf("%s:%d", listen, port),
				Handler:           handler,
				ReadHeaderTimeout: 500 * time.Millisecond,
			}

			return server.ListenAndServe()
		},
	}
)

type MercuryV02Response struct {
	ChainlinkBlob string `json:"chainlinkBlob"`
}

type MercuryV03Response struct {
	Reports []MercuryV03Report `json:"reports"`
}

type MercuryV03Report struct {
	FeedID                string `json:"feedID"` // feed id in hex encoded
	ValidFromTimestamp    uint32 `json:"validFromTimestamp"`
	ObservationsTimestamp uint32 `json:"observationsTimestamp"`
	FullReport            string `json:"fullReport"` // the actual hex encoded mercury report of this feed, can be sent to verifier
}

const (
	DefaultMercuryV2Report = `0x` +
		`0001c38d71fed6c320b90e84b6f559459814d068e2a1700adc931ca9717d4fe7` +
		`0000000000000000000000000000000000000000000000000000000001a80b52` +
		`b4bf1233f9cb71144a253a1791b202113c4ab4a92fa1b176d684b4959666ff82` +
		`00000000000000000000000000000000000000000000000000000000000000e0` +
		`0000000000000000000000000000000000000000000000000000000000000200` +
		`0000000000000000000000000000000000000000000000000000000000000260` +
		`0000000000000000000000000000000000000000000000000000000000000000` +
		`0000000000000000000000000000000000000000000000000000000000000100` +
		`4254432d5553442d415242495452554d2d544553544e45540000000000000000` +
		`00000000000000000000000000000000000000000000000000000000645570be` +
		`000000000000000000000000000000000000000000000000000002af2b818dc5` +
		`000000000000000000000000000000000000000000000000000002af2426faf3` +
		`000000000000000000000000000000000000000000000000000002af32dc2097` +
		`00000000000000000000000000000000000000000000000000000000012130f8` +
		`df0a9745bb6ad5e2df605e158ba8ad8a33ef8a0acf9851f0f01668a3a3f2b686` +
		`00000000000000000000000000000000000000000000000000000000012130f6` +
		`0000000000000000000000000000000000000000000000000000000000000002` +
		`c4a7958dce105089cf5edb68dad7dcfe8618d7784eb397f97d5a5fade78c11a5` +
		`8275aebda478968e545f7e3657aba9dcbe8d44605e4c6fde3e24edd5e22c9427` +
		`0000000000000000000000000000000000000000000000000000000000000002` +
		`459c12d33986018a8959566d145225f0c4a4e61a9a3f50361ccff397899314f0` +
		`018162cf10cd89897635a0bb62a822355bd199d09f4abe76e4d05261bb44733d`
)
