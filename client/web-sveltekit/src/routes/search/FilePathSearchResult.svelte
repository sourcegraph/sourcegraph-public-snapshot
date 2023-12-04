<svelte:options immutable />

<script lang="ts">
    import type { PathMatch } from '$lib/shared'
    import CodeHostIcon from './CodeHostIcon.svelte'
    import FileSearchResultHeader from './FileSearchResultHeader.svelte'
    import RepoStars from './RepoStars.svelte'

    import SearchResult from './SearchResult.svelte'

    export let result: PathMatch
</script>

<SearchResult>
    <CodeHostIcon slot="icon" repository={result.repository} />
    <FileSearchResultHeader slot="title" {result} />
    <svelte:fragment slot="info">
        {#if result.repoStars}
            <RepoStars repoStars={result.repoStars} />
        {/if}
    </svelte:fragment>
    <div class="p-2">
        <small>Path match</small>
    </div>
</SearchResult>
