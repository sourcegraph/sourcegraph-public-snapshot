<script lang="ts">
    import type { ComponentProps } from 'svelte'
    import { writable } from 'svelte/store'

    import { getButtonClassName } from '@sourcegraph/wildcard'

    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import { sizeToFit } from '$lib/dom'
    import { registerHotkey } from '$lib/Hotkey'
    import Icon from '$lib/Icon.svelte'
    import GlobalHeaderPortal from '$lib/navigation/GlobalHeaderPortal.svelte'
    import { createScopeSuggestions } from '$lib/search/codemirror/suggestions'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { queryStateStore } from '$lib/search/state'
    import { TELEMETRY_SEARCH_SOURCE_TYPE, repositoryInsertText } from '$lib/shared'
    import { settings } from '$lib/stores'
    import { default as TabsHeader } from '$lib/TabsHeader.svelte'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import { DropdownMenu, MenuLink } from '$lib/wildcard'

    import type { LayoutData } from './$types'
    import { setRepositoryPageContext, type RepositoryPageContext } from './context'
    import RepoMenu from './RepoMenu.svelte'

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
        icon?: ComponentProps<Icon>['icon']
        /**
         * Who can see this entry.
         */
        visibility: 'admin' | 'user'
    }

    export let data: LayoutData

    const menuOpen = writable(false)
    const navEntries: MenuEntry[] = [
        { path: '', icon: ILucideCode, label: 'Code', visibility: 'user' },
        { path: '/-/commits', icon: ILucideGitCommitVertical, label: 'Commits', visibility: 'user' },
        { path: '/-/branches', icon: ILucideGitBranch, label: 'Branches', visibility: 'user' },
        { path: '/-/tags', icon: ILucideTag, label: 'Tags', visibility: 'user' },
        { path: '/-/stats/contributors', icon: ILucideUsers, label: 'Contributors', visibility: 'user' },
    ]
    const menuEntries: MenuEntry[] = [
        { path: '/-/compare', icon: ILucideHistory, label: 'Compare', visibility: 'user' },
        { path: '/-/own', icon: ILucideUsers, label: 'Ownership', visibility: 'admin' },
        { path: '/-/embeddings', icon: ILucideSpline, label: 'Embeddings', visibility: 'admin' },
        { path: '/-/code-graph', icon: ILucideBrainCircuit, label: 'Code graph data', visibility: 'admin' },
        { path: '/-/batch-changes', icon: ISgBatchChanges, label: 'Batch changes', visibility: 'admin' },
        { path: '/-/settings', icon: ILucideSettings, label: 'Settings', visibility: 'admin' },
    ]
    const repositoryContext = writable<RepositoryPageContext>({})
    const contextSearchSuggestions = createScopeSuggestions({
        getContextInformation() {
            return {
                repoName: data.repoName,
                revision: $repositoryContext.revision ?? data.displayRevision,
                directoryPath: $repositoryContext.directoryPath,
                filePath: $repositoryContext.filePath,
                fileLanguage: $repositoryContext.fileLanguage,
            }
        },
    })

    setRepositoryPageContext(repositoryContext)

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

    $: ({ repoName, revision } = data)
    $: query = `repo:${repositoryInsertText({ repository: repoName })}${revision ? `@${revision}` : ''} `
    $: queryState = queryStateStore({ query }, $settings)
    function handleSearchSubmit(): void {
        TELEMETRY_RECORDER.recordEvent('search', 'submit', {
            metadata: { source: TELEMETRY_SEARCH_SOURCE_TYPE['repo'] },
        })
    }

    registerHotkey({
        keys: {
            key: 'ctrl+backspace',
            mac: 'cmd+backspace',
        },
        ignoreInputFields: false,
        handler: () => {
            goto(data.repoURL)
        },
    })
</script>

<GlobalHeaderPortal>
    <div class="search-header">
        <SearchInput {queryState} size="compat" onSubmit={handleSearchSubmit} extension={contextSearchSuggestions} />
    </div>
</GlobalHeaderPortal>

<nav
    aria-label="repository"
    use:sizeToFit={{
        grow() {
            visibleNavEntryCount = Math.min(visibleNavEntryCount + 1, viewableNavEntries.length)
            return visibleNavEntryCount < viewableNavEntries.length
        },
        shrink() {
            visibleNavEntryCount = Math.max(visibleNavEntryCount - 1, 0)
            return visibleNavEntryCount > 0
        },
    }}
>
    <RepoMenu
        repoName={data.repoName}
        displayRepoName={data.displayRepoName}
        repoURL={data.repoURL}
        externalURL={data.resolvedRevision?.repo?.externalURLs?.[0].url}
        externalServiceKind={data.resolvedRevision?.repo?.externalURLs?.[0].serviceKind ?? undefined}
    />

    <TabsHeader id="repoheader" {tabs} selected={selectedTab} />

    <DropdownMenu
        open={menuOpen}
        triggerButtonClass={getButtonClassName({ variant: 'icon', outline: true, size: 'sm' })}
        aria-label="{$menuOpen ? 'Close' : 'Open'} repo navigation"
    >
        <svelte:fragment slot="trigger">
            <Icon icon={ILucideEllipsis} aria-label="More repo navigation items" />
        </svelte:fragment>
        {#each allMenuEntries as entry}
            {#if entry.visibility === 'user' || (entry.visibility === 'admin' && data.user?.siteAdmin)}
                {@const href = data.repoURL + entry.path}
                <MenuLink {href}>
                    <div class="overflow-entry">
                        {#if entry.icon}
                            <Icon icon={entry.icon} inline aria-hidden />
                        {/if}
                        <span>{entry.label}</span>
                    </div>
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
    }

    .overflow-entry {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
</style>
