<script lang="ts">
    import { mdiAccount, mdiCodeTags, mdiCog, mdiHistory, mdiSourceBranch, mdiSourceCommit, mdiTag } from '@mdi/js'
    import { writable } from 'svelte/store'

    import { getButtonClassName } from '@sourcegraph/wildcard'

    import { page } from '$app/stores'
    import { sizeToFit } from '$lib/dom'
    import Icon2 from '$lib/Icon2.svelte'
    import Icon from '$lib/Icon.svelte'
    import GlobalHeaderPortal from '$lib/navigation/GlobalHeaderPortal.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { QueryState, queryStateStore } from '$lib/search/state'
    import { repositoryInsertText } from '$lib/shared'
    import { settings } from '$lib/stores'
    import { default as TabsHeader } from '$lib/TabsHeader.svelte'
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'
    import { TELEMETRY_V2_RECORDER } from '$lib/telemetry2'
    import { DropdownMenu, MenuLink } from '$lib/wildcard'

    import type { LayoutData } from './$types'

    interface MenuEntry {
        /**
         * The (URL) path to the page.
         */
        path: string
        /**
         * The visible name of the menu entry.
         */
        label: string
        /**
         * The icon to display next to the title.
         */
        icon?: string
        /**
         * Who can see this entry.
         */
        visibility: 'admin' | 'user'
    }

    export let data: LayoutData

    const menuOpen = writable(false)
    const navEntries: MenuEntry[] = [
        { path: '', icon: mdiCodeTags, label: 'Code', visibility: 'user' },
        { path: '/-/commits', icon: mdiSourceCommit, label: 'Commits', visibility: 'user' },
        { path: '/-/branches', icon: mdiSourceBranch, label: 'Branches', visibility: 'user' },
        { path: '/-/tags', icon: mdiTag, label: 'Tags', visibility: 'user' },
        { path: '/-/stats/contributors', icon: mdiAccount, label: 'Contributors', visibility: 'user' },
    ]
    const menuEntries: MenuEntry[] = [
        { path: '/-/compare', icon: mdiHistory, label: 'Compare', visibility: 'user' },
        { path: '/-/own', icon: mdiAccount, label: 'Ownership', visibility: 'admin' },
        { path: '/-/embeddings', label: 'Embeddings', visibility: 'admin' },
        { path: '/-/code-graph', label: 'Code graph data', visibility: 'admin' },
        { path: '/-/batch-changes', label: 'Batch changes', visibility: 'admin' },
        { path: '/-/settings', icon: mdiCog, label: 'Settings', visibility: 'admin' },
    ]

    $: viewableNavEntries = navEntries.filter(
        entry => entry.visibility === 'user' || (entry.visibility === 'admin' && data.user?.siteAdmin)
    )
    $: visibleNavEntryCount = viewableNavEntries.length
    $: navEntriesToShow = viewableNavEntries.slice(0, visibleNavEntryCount)
    $: overflowNavEntries = viewableNavEntries.slice(visibleNavEntryCount)
    $: allMenuEntries = [...overflowNavEntries, ...menuEntries]

    function isCodePage(repoURL: string, pathname: string) {
        return (
            pathname === repoURL || pathname.startsWith(`${repoURL}/-/blob`) || pathname.startsWith(`${repoURL}/-/tree`)
        )
    }

    function isActive(href: string, url: URL): boolean {
        return href === data.repoURL ? isCodePage(data.repoURL, $page.url.pathname) : url.pathname.startsWith(href)
    }
    $: tabs = navEntriesToShow.map(entry => ({
        id: entry.label,
        title: entry.label,
        icon: entry.icon,
        href: data.repoURL + entry.path,
    }))
    $: selectedTab = tabs.findIndex(tab => isActive(tab.href, $page.url))

    $: ({ repoName, displayRepoName, revision, resolvedRevision } = data)
    $: query = `repo:${repositoryInsertText({ repository: repoName })}${revision ? `@${revision}` : ''} `
    $: queryState = queryStateStore({ query }, $settings)
    function handleSearchSubmit(state: QueryState): void {
        SVELTE_LOGGER.log(
            SVELTE_TELEMETRY_EVENTS.SearchSubmit,
            { source: 'repo', query: state.query },
            { source: 'repo', patternType: state.patternType }
        )
        TELEMETRY_V2_RECORDER.recordEvent('search', 'submit', {
            metadata: { source: TELEMETRY_V2_SEARCH_SOURCE_TYPE['repo'] },
        })
    }
</script>

<GlobalHeaderPortal>
    <div class="search-header">
        <SearchInput {queryState} size="compat" onSubmit={handleSearchSubmit} />
    </div>
</GlobalHeaderPortal>

<nav
    aria-label="repository"
    use:sizeToFit={{
        grow() {
            visibleNavEntryCount = Math.min(visibleNavEntryCount + 1, viewableNavEntries.length)
            return visibleNavEntryCount === viewableNavEntries.length
        },
        shrink() {
            visibleNavEntryCount = Math.max(visibleNavEntryCount - 1, 0)
            return visibleNavEntryCount === 0
        },
    }}
>
    <a href={data.repoURL}>
        <CodeHostIcon repository={repoName} codeHost={resolvedRevision?.repo?.externalRepository?.serviceType} />
        <h1>{displayRepoName}</h1>
    </a>

    <TabsHeader id="repoheader" {tabs} selected={selectedTab} />

    <DropdownMenu
        open={menuOpen}
        triggerButtonClass={getButtonClassName({ variant: 'icon', outline: true, size: 'sm' })}
        aria-label="{$menuOpen ? 'Close' : 'Open'} repo navigation"
    >
        <svelte:fragment slot="trigger">
            <Icon2 icon={ILucideEllipsis} aria-label="More repo navigation items" />
        </svelte:fragment>
        {#each allMenuEntries as entry}
            {#if entry.visibility === 'user' || (entry.visibility === 'admin' && data.user?.siteAdmin)}
                {@const href = data.repoURL + entry.path}
                <MenuLink {href}>
                    <span class="overflow-entry" class:active={isActive(href, $page.url)}>
                        {#if entry.icon}
                            <Icon svgPath={entry.icon} inline />
                        {/if}
                        <span>{entry.label}</span>
                    </span>
                </MenuLink>
            {/if}
        {/each}
    </DropdownMenu>
</nav>

<slot />

<style lang="scss">
    .search-header {
        width: 100%;
        z-index: 1;
    }

    nav {
        flex: none;

        color: var(--body-color);
        display: flex;
        align-items: stretch;
        justify-items: flex-start;
        gap: 0.5rem;
        overflow: hidden;
        border-bottom: 1px solid var(--border-color);
        background-color: var(--color-bg-1);

        a {
            all: unset;

            display: flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0 1rem;
            cursor: pointer;
            &:hover {
                background-color: var(--color-bg-2);
            }

            h1 {
                display: contents;
                font-size: 1rem;
                white-space: nowrap;
                color: var(--text-title);
                font-weight: normal;
            }
        }

        :global([data-dropdown-trigger]) {
            height: 100%;
            align-self: stretch;
            padding: 0.5rem;
            --icon-fill-color: var(--text-muted);
        }
    }

    .overflow-entry {
        width: 100%;
        display: inline-block;
        padding: 0 0.25rem;
        border-radius: var(--border-radius);
    }
</style>
