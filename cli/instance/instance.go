package instance

import (
	"context"
	"errors"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
)

// CreateInstaceIgnorePlatformIndexErrors creates and return an instance of the
// Arduino Core Engine, but won't stop on platforms index loading errors.
func CreateInstaceIgnorePlatformIndexErrors() *rpc.Instance {
	return initInstance().GetInstance()
}

// CreateInstance creates and return an instance of the Arduino Core engine
func CreateInstance() *rpc.Instance {
	resp := initInstance()
	if resp.GetPlatformsIndexErrors() != nil {
		for _, err := range resp.GetPlatformsIndexErrors() {
			formatter.PrintError(errors.New(err), "Error loading index")
		}
		formatter.PrintErrorMessage("Launch '" + os.Args[0] + " core update-index' to fix or download indexes.")
		os.Exit(errorcodes.ErrGeneric)
	}
	return resp.GetInstance()
}

func initInstance() *rpc.InitResp {
	logrus.Info("Initializing package manager")
	req := packageManagerInitReq()

	resp, err := commands.Init(context.Background(), req, output.ProgressBar(), output.TaskProgress(), globals.HTTPClientHeader)
	if err != nil {
		formatter.PrintError(err, "Error initializing package manager")
		os.Exit(errorcodes.ErrGeneric)
	}
	if resp.GetLibrariesIndexError() != "" {
		commands.UpdateLibrariesIndex(context.Background(),
			&rpc.UpdateLibrariesIndexReq{Instance: resp.GetInstance()}, output.ProgressBar())
		rescResp, err := commands.Rescan(context.Background(), &rpc.RescanReq{Instance: resp.GetInstance()})
		if rescResp.GetLibrariesIndexError() != "" {
			formatter.PrintErrorMessage("Error loading library index: " + rescResp.GetLibrariesIndexError())
			os.Exit(errorcodes.ErrGeneric)
		}
		if err != nil {
			formatter.PrintError(err, "Error loading library index")
			os.Exit(errorcodes.ErrGeneric)
		}
		resp.LibrariesIndexError = rescResp.LibrariesIndexError
		resp.PlatformsIndexErrors = rescResp.PlatformsIndexErrors
	}
	return resp
}

func packageManagerInitReq() *rpc.InitReq {
	urls := []string{}
	for _, URL := range globals.Config.BoardManagerAdditionalUrls {
		urls = append(urls, URL.String())
	}

	conf := &rpc.Configuration{}
	conf.DataDir = globals.Config.DataDir.String()
	conf.DownloadsDir = globals.Config.DownloadsDir().String()
	conf.BoardManagerAdditionalUrls = urls
	if globals.Config.SketchbookDir != nil {
		conf.SketchbookDir = globals.Config.SketchbookDir.String()
	}

	return &rpc.InitReq{Configuration: conf}
}
