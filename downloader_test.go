package downloader_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/qor/downloader"
	"testing"
)

func TestDownloader(t *testing.T) {
	downloader := downloader.New("/Users/bin/Codes/go/src/github.com/qor")
	t.Errorf(color.RedString(fmt.Sprintf("\nDownloader TestCase #%d: Failure (%s)\n", 1, downloader)))
}
