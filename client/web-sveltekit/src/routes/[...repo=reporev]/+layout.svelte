<script lang="ts">
    import {
        mdiAccount,
        mdiCodeTags,
        mdiCog,
        mdiHistory,
        mdiSourceBranch,
        mdiSourceCommit,
        mdiTag,
        mdiDotsHorizontal,
    } from '@mdi/js'
    import { writable } from 'svelte/store'

    import { getButtonClassName } from '@sourcegraph/wildcard'

    import { page } from '$app/stores'
    import { computeFit } from '$lib/dom'
    import Icon from '$lib/Icon.svelte'
    import GlobalHeaderPortal from '$lib/navigation/GlobalHeaderPortal.svelte'
    import { DropdownMenu, MenuLink } from '$lib/wildcard'

    import type { LayoutData } from './$types'
    import RepoSearchInput from './RepoSearchInput.svelte'

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

    let visibleNavEntries = navEntries.length
    $: navEntriesToShow = visibleNavEntries === navEntries.length ? navEntries : navEntries.slice(0, visibleNavEntries)
    $: overflowMenu = visibleNavEntries !== navEntries.length ? navEntries.slice(visibleNavEntries) : []
    $: allMenuEntries = [...overflowMenu, ...menuEntries]

    function isCodePage(repoURL: string, pathname: string) {
        return (
            pathname === repoURL || pathname.startsWith(`${repoURL}/-/blob`) || pathname.startsWith(`${repoURL}/-/tree`)
        )
    }

    function isActive(href: string, url: URL): boolean {
        return href === data.repoURL ? isCodePage(data.repoURL, $page.url.pathname) : url.pathname.startsWith(href)
    }

    $: ({ repoName, displayRepoName } = data)
</script>

<GlobalHeaderPortal>
    <nav aria-label="repository">
        <h1><a href="/{repoName}">{displayRepoName}</a></h1>

        <ul use:computeFit on:fit={event => (visibleNavEntries = event.detail.itemCount)}>
            {#each navEntriesToShow as entry}
                {#if entry.visibility === 'user' || (entry.visibility === 'admin' && data.user?.siteAdmin)}
                    {@const href = data.repoURL + entry.path}
                    <li>
                        <a {href} aria-current={isActive(href, $page.url) ? 'page' : undefined}>
                            {#if entry.icon}
                                <Icon svgPath={entry.icon} inline />
                            {/if}
                            <span>{entry.label}</span>
                        </a>
                    </li>
                {/if}
            {/each}
        </ul>

        <DropdownMenu
            open={menuOpen}
            triggerButtonClass={getButtonClassName({ variant: 'icon', outline: true, size: 'sm' })}
            aria-label="{$menuOpen ? 'Close' : 'Open'} repo navigation"
        >
            <svelte:fragment slot="trigger">
                <Icon svgPath={mdiDotsHorizontal} aria-label="More repo navigation items" />
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
        <RepoSearchInput repoName={data.repoName} revision={data.displayRevision} />
    </nav>
</GlobalHeaderPortal>

<slot />

<style lang="scss">
    nav {
        display: flex;
        align-items: baseline;
        gap: 0.5rem;
        overflow: hidden;
        flex: 1;
        min-width: 0;

        a {
            color: var(--text-body);
            text-decoration: none;
        }

        :global([data-dropdown-trigger]) {
            height: 100%;
            align-self: stretch;
            padding: 0.5rem;
            fill: var(--icon-color);
        }
    }

    h1 {
        margin: 0 1rem 0 0;
        font-size: 1rem;
        white-space: nowrap;

        a {
            color: var(--text-title);
        }
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        gap: 0.5rem;
        overflow: hidden;
        align-self: center;
        flex: 1;

        li a {
            display: flex;
            height: 100%;
            align-items: center;
            padding: 0.25rem 0.5rem;
            border-radius: var(--border-radius);
            white-space: nowrap;
            gap: 0.25rem;

            &:hover {
                background-color: var(--color-bg-2);
            }

            &[aria-current='page'] {
                background-color: var(--color-bg-3);
                color: var(--text-title);
            }
        }

        :global([data-icon]) {
            --color: var(--icon-color);
        }
    }

    .overflow-entry {
        width: 100%;
        display: inline-block;
        padding: 0 0.25rem;
        border-radius: var(--border-radius);
    }

    .active {
        background-color: var(--color-bg-3);
    }

    nav {
        color: var(--body-color);
    }
</style>
