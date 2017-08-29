import * as querystring from 'querystring';
import * as React from 'react';
import { render } from 'react-dom';
import { handleSearchInput } from 'sourcegraph/search';
import { AdvancedSearchDrawer } from 'sourcegraph/search/AdvancedSearchDrawer';
import { SearchBox } from 'sourcegraph/search/SearchBox';
import { SearchResults } from 'sourcegraph/search/SearchResults';
import { setState as setSearchState, store as searchStore } from 'sourcegraph/search/store';
import * as URI from 'urijs';

export function injectSearchBox(): void {
    const widget = document.getElementById('search-widget') as HTMLElement;
    if (widget) {
        render(<SearchBox />, widget);
    }
}

export function injectSearchResults(): void {
    const widget = document.getElementById('search-results-widget') as HTMLElement;
    if (widget) {
        render(<SearchResults />, widget);
    }
}

export function injectSearchInputHandler(): void {
    const input = document.getElementById('search-input') as HTMLInputElement;
    const urlQuery = querystring.parse(URI.parse(window.location.href).query);
    if (input) {
        input.value = urlQuery.q || '';
        input.addEventListener('keydown', e => {
            const params = { ...searchStore.getValue(), query: (e.target as any).value };
            setSearchState(params);
            handleSearchInput(e, params);
        });
    }
}

export function injectAdvancedSearchToggle(): void {
    const searchBoxContainer = document.getElementById('search-box-container') as HTMLElement;
    render(<SearchBox />, searchBoxContainer);
}

export function injectAdvancedSearchDrawer(): void {
    const el = document.querySelector('#advanced-search') as HTMLElement;
    searchStore.subscribe(state => {
        el.style.display = state.showAdvancedSearch ? 'block' : 'none';
    });
    render(<AdvancedSearchDrawer />, el);
}
