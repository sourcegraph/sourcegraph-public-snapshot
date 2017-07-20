import { injectReferencesWidget } from "app/references/inject";
import { setReferences } from "app/references/store";

document.addEventListener("DOMContentLoaded", () => {
	injectReferencesWidget();
	//do work
	setReferences({
		docked: true,
		context: {
			path: "mux.go",
			repoRevSpec: {
				repoURI: "github.com/gorilla/mux",
				rev: "ac112f7d75a0714af1bd86ab17749b31f7809640",
				isDelta: false,
				isBase: false,
			},
			coords: {
				line: 40,
				char: 26,
				word: "Handler",
			},
		},
	});
});
