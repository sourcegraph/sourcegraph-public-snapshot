<svelte:options immutable />

<script lang="ts">
    import RepoStars from '$lib/repo/RepoStars.svelte'
    import type { PathMatch } from '$lib/shared'

    import FileSearchResultHeader from './FileSearchResultHeader.svelte'
    import PreviewButton from './PreviewButton.svelte'
    import SearchResult from './SearchResult.svelte'

    export let result: PathMatch
</script>

<SearchResult>
    <FileSearchResultHeader slot="title" {result} />
    <svelte:fragment slot="info">
        {#if result.repoStars}
            <RepoStars repoStars={result.repoStars} />
        {/if}
        <PreviewButton {result} />
    </svelte:fragment>
</SearchResult>
