// +build js

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

func init() {
	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) { setup2() })
}

func setup2() {
	commentEditor, ok := document.GetElementByID("comment-editor").(*dom.HTMLTextAreaElement)
	if !ok {
		return
	}

	commentEditor.AddEventListener("paste", false, pasteHandler)
}

func pasteHandler(e dom.Event) {
	ce := e.(*dom.ClipboardEvent)

	items := ce.Get("clipboardData").Get("items")
	if items.Length() == 0 {
		return
	}
	item := items.Index(0)
	if item.Get("kind").String() != "file" {
		return
	}
	if item.Get("type").String() != "image/png" {
		return
	}
	file := item.Call("getAsFile")

	go func() {
		b := blobToBytes(file)

		resp, err := http.Post("/ui/.usercontent", "image/png", bytes.NewReader(b))
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		var upload struct {
			Name  string
			Error string
		}
		err = json.NewDecoder(resp.Body).Decode(&upload)
		if err != nil {
			log.Println(err)
			return
		}
		if upload.Error != "" {
			log.Println(upload.Error)
			return
		}

		url := "/usercontent/" + upload.Name
		insertText(ce.Target().(*dom.HTMLTextAreaElement), "![Image]("+url+")\n")
	}()
}

func insertText(t *dom.HTMLTextAreaElement, inserted string) {
	value, start, end := t.Value, t.SelectionStart, t.SelectionEnd
	t.Value = value[:start] + inserted + value[end:]
	t.SelectionStart, t.SelectionEnd = start+len(inserted), start+len(inserted)
}

// blobToBytes converts a Blob to []byte.
func blobToBytes(blob *js.Object) []byte {
	var b = make(chan []byte)
	fileReader := js.Global.Get("FileReader").New()
	fileReader.Set("onload", func() {
		b <- js.Global.Get("Uint8Array").New(fileReader.Get("result")).Interface().([]byte)
	})
	fileReader.Call("readAsArrayBuffer", blob)
	return <-b
}
