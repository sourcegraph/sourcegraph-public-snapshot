<script lang="ts">
    import { getRepoMatchLabel, getRepoMatchUrl, type RepositoryMatch } from '$lib/shared'

    import SearchResult from './SearchResult.svelte'

    export let result: RepositoryMatch

    $: repoName = getRepoMatchLabel(result)
    $: repoAtRevisionURL = getRepoMatchUrl(result)
</script>

<SearchResult {result}>
    <div slot="title">
        <a href={repoAtRevisionURL}>{repoName}</a>
    </div>
</SearchResult>
