package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
)

func main() {
	var opts struct {
		Folder string `short:"f" long:"folder" description:"Folder containing PHP extension metadata files" required:"true"`
	}

	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	depFactory := dependency.NewDependencyFactory("")
	webClient := utils.NewPHPExtensionsWebClient()
	checkSummer := utils.Checksummer{}

	phpExtensionsUtils := utils.NewPHPExtensionsUtils(depFactory, webClient, checkSummer)
	results, err := phpExtensionsUtils.GenerateJSONPayload(opts.Folder)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(results)
}
