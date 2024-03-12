<script lang="ts">
    import type { ContentMatch, PathMatch, SymbolMatch } from '$lib/shared'

    import { getSearchResultsContext } from './searchResultsContext'

    export let result: ContentMatch | SymbolMatch | PathMatch

    const searchResultContext = getSearchResultsContext()
</script>

<button on:click={() => searchResultContext.setPreview(result)}>Preview</button>
