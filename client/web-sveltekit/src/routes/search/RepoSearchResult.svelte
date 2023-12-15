<svelte:options immutable />

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { featureFlag } from '$lib/featureflags'
    import { displayRepoName, getRepoMatchUrl, type RepositoryMatch } from '$lib/shared'
    import { mdiArchive, mdiLock, mdiSourceFork } from '@mdi/js'
    import CodeHostIcon from './CodeHostIcon.svelte'

    import SearchResult from './SearchResult.svelte'
    import { Badge } from '$lib/wildcard'
    import { getSearchResultsContext } from './searchResultsContext'
    import { limitDescription, getMetadata, buildSearchURLQueryForMeta, simplifyLineRange } from '$lib/search/results'
    import { highlightRanges } from '$lib/dom'

    export let result: RepositoryMatch

    const enableRepositoryMetadata = featureFlag('repository-metadata')
    const queryState = getSearchResultsContext().queryState

    $: repoAtRevisionURL = getRepoMatchUrl(result)
    $: metadata = $enableRepositoryMetadata ? getMetadata(result) : []
    $: description = limitDescription(result.description ?? '')
    $: repoName = displayRepoName(result.repository)

    $: repositoryMatches = result.repositoryMatches?.map(simplifyLineRange) ?? []
    $: if (repoName !== result.repository) {
        // We only display part of the repository name, therefore we have to
        // adjust the match ranges for highlighting
        const delta = result.repository.length - repoName.length
        repositoryMatches = repositoryMatches.map(([start, end]) => [start - delta, end - delta])
    }
    $: descriptionMatches = result.descriptionMatches?.map(simplifyLineRange) ?? []
</script>

<SearchResult>
    <CodeHostIcon slot="icon" repository={result.repository} />
    <div slot="title">
        <!-- #key is needed here to recreate the link because use:highlightRanges changes the DOM -->
        {#key repositoryMatches}
            <a href={repoAtRevisionURL} use:highlightRanges={{ ranges: repositoryMatches }}
                >{displayRepoName(result.repository)}</a
            >
        {/key}
        {#if result.fork}
            <span class="info">
                <Icon aria-label="Forked repository" svgPath={mdiSourceFork} inline />
                <small>Fork</small>
            </span>
        {/if}
        {#if result.archived}
            <span class="info">
                <Icon aria-label="Archived repository" svgPath={mdiArchive} inline />
                <small>Archive</small>
            </span>
        {/if}
        {#if result.private}
            <span class="info">
                <Icon aria-label="Private repository" svgPath={mdiLock} inline />
                <small>Private</small>
            </span>
        {/if}
    </div>
    {#if description}
        <!-- #key is needed here to recreate the paragraph because use:highlightRanges changes the DOM -->
        {#key description}
            <p class="p-2 m-0" use:highlightRanges={{ ranges: descriptionMatches }}>
                {limitDescription(description)}
            </p>
        {/key}
    {/if}
    {#if metadata.length > 0}
        <ul class="p-2">
            {#each metadata as meta}
                <li>
                    <Badge variant="outlineSecondary">
                        <a
                            slot="custom"
                            let:class={className}
                            class={className}
                            href="/search?{buildSearchURLQueryForMeta($queryState, meta)}"
                        >
                            <code
                                >{meta.key}{#if meta.value}:{meta.value}{/if}</code
                            >
                        </a>
                    </Badge>
                </li>
            {/each}
        </ul>
    {/if}
</SearchResult>

<style lang="scss">
    ul {
        margin: 0;
        list-style: none;
        display: flex;
        gap: 0.5rem;
        flex-wrap: wrap;

        code {
            color: var(--search-filter-keyword-color);
        }
    }

    .info {
        border-left: 1px solid var(--border-color);
        margin-left: 0.5rem;
        padding-left: 0.5rem;
    }
</style>
