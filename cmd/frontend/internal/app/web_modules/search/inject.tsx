import { SearchForm } from "app/search/SearchForm";
import { SearchResults } from "app/search/SearchResults";
import * as React from "react";
import { render } from "react-dom";

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
