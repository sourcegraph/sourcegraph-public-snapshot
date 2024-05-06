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
    import { page } from '$app/stores'
    import { getButtonClassName } from '@sourcegraph/wildcard'

    import { computeFit } from '$lib/dom'
    import { DropdownMenu, MenuLink } from '$lib/wildcard'
    import Icon from '$lib/Icon.svelte'
    import GlobalHeaderPortal from '$lib/navigation/GlobalHeaderPortal.svelte'

    import type { LayoutData } from './$types'
    import RepoSearchInput from './RepoSearchInput.svelte'

    export let data: LayoutData

    const menuOpen = writable(false)
    const navEntries: { path: string; icon: string; title: string }[] = [
        { path: '', icon: mdiCodeTags, title: 'Code' },
        { path: '/-/commits', icon: mdiSourceCommit, title: 'Commits' },
        { path: '/-/branches', icon: mdiSourceBranch, title: 'Branches' },
        { path: '/-/tags', icon: mdiTag, title: 'Tags' },
        { path: '/-/stats/contributors', icon: mdiAccount, title: 'Contributors' },
    ]
    const menuEntries: { path: string; icon: string; title: string }[] = [
        { path: '/-/compare', icon: mdiHistory, title: 'Compare' },
        { path: '/-/own', icon: mdiAccount, title: 'Ownership' },
        { path: '/-/embeddings', icon: '', title: 'Embeddings' },
        { path: '/-/batch-changes', icon: '', title: 'Batch changes' },
        { path: '/-/settings', icon: mdiCog, title: 'Settings' },
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
                {@const href = data.repoURL + entry.path}
                <li>
                    <a {href} aria-current={isActive(href, $page.url) ? 'page' : undefined}>
                        {#if entry.icon}
                            <Icon svgPath={entry.icon} inline />
                        {/if}
                        <span>{entry.title}</span>
                    </a>
                </li>
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
                {@const href = data.repoURL + entry.path}
                <MenuLink {href}>
                    <span class="overflow-entry" class:active={isActive(href, $page.url)}>
                        {#if entry.icon}
                            <Icon svgPath={entry.icon} inline />
                        {/if}
                        <span>{entry.title}</span>
                    </span>
                </MenuLink>
            {/each}
        </DropdownMenu>
        <RepoSearchInput repoName={data.repoName} />
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
