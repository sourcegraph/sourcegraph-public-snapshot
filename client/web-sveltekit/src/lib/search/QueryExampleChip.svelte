<script context="module" lang="ts">
    export interface QueryExample {
        query: string
        id?: string
        slug?: string
        helperText?: string
    }
</script>

<script lang="ts">
    import { getContext } from 'svelte'

    import SyntaxHighlightedQuery from '$lib/search/SyntaxHighlightedQuery.svelte'

    import type { SearchPageContext } from './utils'

    export let queryExample: QueryExample

    const searchContext = getContext<SearchPageContext>('search-context')
    function handleClick() {
        searchContext.setQuery(query => (query + ' ' + queryExample.query).trim())
    }
</script>

<li class="d-flex align-items-center">
    <button type="button" on:click={handleClick}>
        <SyntaxHighlightedQuery query={queryExample.query} />
    </button>
    {#if queryExample.helperText}
        <span class="text-muted ml-2"><small>{queryExample.helperText}</small></span>
    {/if}
</li>

<style lang="scss">
    button {
        background-color: var(--code-bg);
        box-shadow: var(--search-input-shadow);
        border-radius: var(--border-radius);
        padding: 0.125rem 0.375rem;
        font-size: 0.75rem;
        max-width: 21rem;
        text-align: left;
        border: 1px solid transparent;
        color: var(--body-color);
        cursor: pointer;

        &:hover {
            border: 1px solid var(--border-color);
        }

        &:active {
            position: relative;
            top: 1px;
            border: 1px solid var(--border-color);
            box-shadow: none;
        }

        &:focus {
            border: 1px solid transparent;
            box-shadow: 0 0 0 0.125rem var(--primary-2);
            outline: none;
        }

        &:active:focus {
            border: 1px solid var(--border-color);
            box-shadow: none;
        }
    }
</style>
