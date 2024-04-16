<svelte:options immutable />

<script lang="ts">
    import { mdiArchive, mdiLock, mdiSourceFork } from '@mdi/js'

    import { highlightRanges } from '$lib/dom'
    import { featureFlag } from '$lib/featureflags'
    import Icon from '$lib/Icon.svelte'
    import RepoStars from '$lib/repo/RepoStars.svelte'
    import { limitDescription, getRepositoryBadges, simplifyLineRange } from '$lib/search/results'
    import type { RepositoryMatch } from '$lib/shared'
    import { Badge } from '$lib/wildcard'

    import RepoRev from './RepoRev.svelte'
    import SearchResult from './SearchResult.svelte'
    import { getSearchResultsContext } from './searchResultsContext'

    export let result: RepositoryMatch

    const enableRepositoryMetadata = featureFlag('repository-metadata')
    const queryState = getSearchResultsContext().queryState

    $: badges = getRepositoryBadges($queryState, result, $enableRepositoryMetadata)
    $: description = limitDescription(result.description ?? '')
    $: repositoryMatches = result.repositoryMatches?.map(simplifyLineRange) ?? []
    $: descriptionMatches = result.descriptionMatches?.map(simplifyLineRange) ?? []
    $: rev = result.branches?.[0]
</script>

<SearchResult>
    <div slot="title">
        <RepoRev repoName={result.repository} {rev} highlights={repositoryMatches} />
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
    <svelte:fragment slot="info">
        {#if result.repoStars}
            <RepoStars repoStars={result.repoStars} />
        {/if}
    </svelte:fragment>
    {#if description}
        <!-- #key is needed here to recreate the paragraph because use:highlightRanges changes the DOM -->
        {#key description}
            <p class="p-2 m-0" use:highlightRanges={{ ranges: descriptionMatches }}>
                {limitDescription(description)}
            </p>
        {/key}
    {/if}<!--
        Intentional weird comment to avoid adding an empty line to the body
    -->{#if badges.length > 0}
        <ul class="p-2">
            {#each badges as badge}
                <li>
                    <Badge variant="outlineSecondary">
                        <a slot="custom" let:class={className} class={className} href={`/search?${badge.urlQuery}`}>
                            <code>{badge.label}</code>
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
