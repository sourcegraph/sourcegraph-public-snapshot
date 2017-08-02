import { handleSearchInput } from "app/search";
import { SearchForm } from "app/search/SearchForm";
import { SearchResults } from "app/search/SearchResults";
import * as querystring from "querystring";
import * as React from "react";
import { render } from "react-dom";
import * as URI from "urijs";

export function injectSearchForm(): void {
	const widget = document.getElementById("search-widget") as HTMLElement;
	if (widget) {
		render(<SearchForm />, widget);
	}
}

export function injectSearchResults(): void {
	const widget = document.getElementById("search-results-widget") as HTMLElement;
	if (widget) {
		render(<SearchResults />, widget);
	}
}

export function injectSearchInputHandler(): void {
	const input = document.getElementById("search-input") as HTMLInputElement;
	const urlQuery = querystring.parse(URI.parse(window.location.href).query);
	if (input) {
		input.value = urlQuery["q"] || "";
		input.addEventListener("keydown", (e) => handleSearchInput(e, true));
	}
}
