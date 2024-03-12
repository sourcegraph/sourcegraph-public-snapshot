<script lang="ts">
    import type { ContentMatch, PathMatch, SymbolMatch } from '$lib/shared'

    import { getSearchResultsContext } from './searchResultsContext'

    export let result: ContentMatch | SymbolMatch | PathMatch

    const searchResultContext = getSearchResultsContext()
</script>

<button on:click={() => searchResultContext.setPreview(result)}>Preview</button>

<style lang="scss">
    button {
        border: none;
        background-color: transparent;
        font-weight: 500;
        color: var(--link-color);
        &:hover {
            text-decoration: underline;
        }
    }
</style>
