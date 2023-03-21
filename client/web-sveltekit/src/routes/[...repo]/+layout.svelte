<script lang="ts">
    import { mdiAccount, mdiCodeTags, mdiSourceBranch, mdiSourceCommit, mdiSourceRepository, mdiTag } from '@mdi/js'
    import { setContext } from 'svelte'

    import { page } from '$app/stores'
    import { isErrorLike } from '$lib/common'
    import { displayRepoName, isRepoNotFoundErrorLike } from '$lib/shared'
    import { createActionStore, type ActionStore } from '$lib/repo/actions'
    import { getRevisionLabel, navFromPath } from '$lib/repo/utils'
    import Icon from '$lib/Icon.svelte'

    import RepoNotFoundError from './RepoNotFoundError.svelte'
    import Permalink from './Permalink.svelte'
    import type { LayoutData } from './$types'

    export let data: LayoutData

    const menu = [
        { path: '/-/commits', icon: mdiSourceCommit, title: 'Commits' },
        { path: '/-/branches', icon: mdiSourceBranch, title: 'Branches' },
        { path: '/-/tags', icon: mdiTag, title: 'Tags' },
        { path: '/-/stats/contributors', icon: mdiAccount, title: 'Contributors' },
    ] as const

    function isCodePage(repoURL: string, pathname: string) {
        return (
            pathname === repoURL || pathname.startsWith(`${repoURL}/-/blob`) || pathname.startsWith(`${repoURL}/-/tree`)
        )
    }

    // Sets up a context for other components to add add buttons to the header
    const repoActions = createActionStore()
    setContext<ActionStore>('repo-actions', repoActions)

    $: viewerCanAdminister = data.user?.siteAdmin ?? false
    $: ({ repo, path } = $page.params)
    $: nav = path ? navFromPath(path, repo, $page.url.pathname.includes('/-/blob/')) : []

    $: resolvedRevision = isErrorLike(data.resolvedRevision) ? null : data.resolvedRevision
    $: revisionLabel = getRevisionLabel(data.revision, resolvedRevision)
    $: repoName = displayRepoName(repo.split('@')[0])
    $: if (resolvedRevision) {
        repoActions.setAction({ key: 'permalink', priority: 100, component: Permalink })
    }
</script>

{#if isErrorLike(data.resolvedRevision)}
    <!--
        We are rendering the error page here instead of using SvelteKit's error handler.
        See comment in +layout.ts
    -->
    {#if isRepoNotFoundErrorLike(data.resolvedRevision)}
        <RepoNotFoundError {repoName} {viewerCanAdminister} />
    {:else}
        Something went wrong
    {/if}
{:else}
    <div class="header">
        <nav>
            <h1><a href="/{repo}"><Icon svgPath={mdiSourceRepository} inline /> {repoName}</a></h1>
            <!--
                TODO: Add back revision
                {#if revisionLabel}
                    @ <span class="button">{revisionLabel}</span>
                {/if}
                -->
            <ul class="menu">
                <li>
                    <a href={data.repoURL} class:active={isCodePage(data.repoURL, $page.url.pathname)}>
                        <Icon svgPath={mdiCodeTags} inline /> <span class="ml-1">Code</span>
                    </a>
                </li>
                {#each menu as entry}
                    {@const href = data.repoURL + entry.path}
                    <li>
                        <a {href} class:active={$page.url.pathname.startsWith(href)}>
                            <Icon svgPath={entry.icon} inline /> <span class="ml-1">{entry.title}</span>
                        </a>
                    </li>
                {/each}
            </ul>
        </nav>

        <div class="actions">
            {#each $repoActions as action (action.key)}
                <svelte:component this={action.component} />
            {/each}
        </div>
    </div>
    <div class="ml-3 mt-1">
        {#if nav.length > 0}
            <span class="crumps">
                {#each nav as [label, url]}
                    <span>/</span>
                    <a href={url}>{label}</a>&nbsp;
                {/each}
            </span>
        {/if}
    </div>
    <slot />
{/if}

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem 1rem;
        border-bottom: 1px solid var(--border-color);
    }

    h1 {
        margin: 0;
        margin-right: 1rem;
        font-size: 1.3rem;
    }

    nav {
        display: flex;
        align-items: center;

        a {
            color: var(--body-color);
            text-decoration: none;
        }
    }

    ul.menu {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;

        li a {
            display: flex;
            height: 100%;
            align-items: center;
            padding: 0.25rem 0.5rem;
            margin: 0 0.25rem;
            border-radius: var(--border-radius);

            &:hover {
                background-color: var(--color-bg-2);
            }

            &.active {
                background-color: var(--color-bg-3);
            }
        }
    }

    .actions {
        margin-left: auto;
    }

    nav {
        color: var(--body-color);
    }

    .crumps {
        color: var(--link-color);
    }

    .button {
        color: var(--body-color);
        border: 1px solid var(--border-color);
        padding: 0.25rem 0.5rem;
        border-radius: var(--border-radius);
        text-decoration: none;
    }
</style>
