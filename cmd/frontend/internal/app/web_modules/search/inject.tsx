import { handleSearchInput } from "app/search";
import { AdvancedSearchDrawer } from "app/search/AdvancedSearchDrawer";
import { AdvancedSearchToggle } from "app/search/AdvancedSearchToggle";
import { SearchForm } from "app/search/SearchForm";
import { SearchResults } from "app/search/SearchResults";
import { setState as setSearchState, store as searchStore } from "app/search/store";
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
		input.addEventListener("keydown", (e) => {
			const params = { ...searchStore.getValue(), query: (e.target as any).value };
			setSearchState(params);
			handleSearchInput(e, params);
		});
	}
}

export function injectAdvancedSearchToggle(): void {
	const el = document.createElement("div");
	el.id = "advanced-search-toggle";
	render(<AdvancedSearchToggle />, el);
	const header = document.querySelector(".header") as HTMLElement;
	header.insertBefore(el, header.querySelector(".fill")!);
}

export function injectAdvancedSearchDrawer(): void {
	const el = document.querySelector("#advanced-search") as HTMLElement;
	searchStore.subscribe((state) => {
		el.style.display = state.showAdvancedSearch ? "block" : "none";
	});
	render(<AdvancedSearchDrawer />, el);
}
