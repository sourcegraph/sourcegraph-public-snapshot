// +build js

package main

import (
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

// UpdateGoPackage updates Go packages specified by importPathPattern.
func UpdateGoPackage(importPathPattern string) {
	var goPackage = document.GetElementByID(importPathPattern)
	var goPackageButton = goPackage.GetElementsByClassName("update-button")[0].(*dom.HTMLAnchorElement)

	goPackageButton.SetTextContent("Updating...")
	goPackageButton.AddEventListener("click", false, func(event dom.Event) { event.PreventDefault() })
	goPackageButton.SetTabIndex(-1)
	goPackageButton.Class().Add("disabled")

	go func() {
		req := xhr.NewRequest("POST", "/-/update")
		req.SetRequestHeader("Content-Type", "application/x-www-form-urlencoded")
		err := req.Send(url.Values{"import_path_pattern": {importPathPattern}}.Encode())
		if err != nil {
			println(err.Error())
			return
		}

		// Hide the "Updating..." label.
		goPackageButton.Style().SetProperty("display", "none", "")

		// Show "No Updates Available" if there are no remaining updates.
		if !hasUpdatesAvailable() {
			document.GetElementByID("no_updates").(dom.HTMLElement).Style().SetProperty("display", "none", "")
		}

		// Move this Go package to "Installed Updates" list.
		installedUpdates := document.GetElementByID("installed_updates").(dom.HTMLElement)
		installedUpdates.Style().SetProperty("display", "", "")
		installedUpdates.ParentNode().InsertBefore(goPackage, installedUpdates.NextSibling()) // Insert after.
	}()
}

// hasUpdatesAvailable returns true if there's at least one remaining update.
func hasUpdatesAvailable() bool {
	updates := document.GetElementsByClassName("go-package-update")
	for _, update := range updates {
		if len(update.GetElementsByClassName("disabled")) == 0 {
			return true
		}
	}
	return false
}

func main() {
	js.Global.Set("update_go_package", UpdateGoPackage)
}
