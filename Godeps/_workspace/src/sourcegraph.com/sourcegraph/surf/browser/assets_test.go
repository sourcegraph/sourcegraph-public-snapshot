package browser

import (
	"bytes"
	"github.com/headzoo/ut"
	"net/url"
	"testing"
)

func TestDownload(t *testing.T) {
	ut.Run(t)

	out := &bytes.Buffer{}
	u, _ := url.Parse("http://i.imgur.com/HW4bJtY.jpg")
	asset := NewImageAsset(u, "", "", "")
	l, err := DownloadAsset(asset, out)
	ut.AssertNil(err)
	ut.AssertGreaterThan(0, int(l))
	ut.AssertEquals(int(l), out.Len())
}

func TestDownloadAsync(t *testing.T) {
	ut.Run(t)

	ch := make(AsyncDownloadChannel, 1)
	u1, _ := url.Parse("http://i.imgur.com/HW4bJtY.jpg")
	u2, _ := url.Parse("http://i.imgur.com/HkPOzEH.jpg")
	asset1 := NewImageAsset(u1, "", "", "")
	asset2 := NewImageAsset(u2, "", "", "")
	out1 := &bytes.Buffer{}
	out2 := &bytes.Buffer{}

	queue := 2
	DownloadAssetAsync(asset1, out1, ch)
	DownloadAssetAsync(asset2, out2, ch)

	for {
		select {
		case result := <-ch:
			ut.AssertGreaterThan(0, int(result.Size))
			if result.Asset == asset1 {
				ut.AssertEquals(int(result.Size), out1.Len())
			} else if result.Asset == asset2 {
				ut.AssertEquals(int(result.Size), out2.Len())
			} else {
				t.Failed()
			}
			queue--
			if queue == 0 {
				goto COMPLETE
			}
		}
	}

COMPLETE:
	close(ch)
	ut.AssertEquals(0, queue)
}
