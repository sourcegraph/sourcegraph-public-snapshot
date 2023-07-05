<script lang="ts">
    import { mdiAccount, mdiCodeTags, mdiSourceBranch, mdiSourceCommit, mdiSourceRepository, mdiTag } from '@mdi/js'
    import { setContext } from 'svelte'

    import { page } from '$app/stores'
    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { createActionStore, type ActionStore } from '$lib/repo/actions'
    import { displayRepoName, isRepoNotFoundErrorLike } from '$lib/shared'

    import type { LayoutData } from './$types'
    import RepoNotFoundError from './RepoNotFoundError.svelte'
    import Header from '$lib/Header.svelte'
    import RevisionSelector from '$lib/repo/ui/RevisionSelector.svelte'
    import { resolvePath } from '@sveltejs/kit'

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
    $: ({ repo } = $page.params)

    function createRevisionURL(revision:string) {
        if ($page.route.id) {
            return resolvePath($page.route.id, {
                repo: `${data.repoName}@${revision}`,
                path: $page.params.path,
            })
        }
        return ''
    }
</script>

{#if isErrorLike(data.resolvedRevisionOrError)}
    <!--
        We are rendering the error page here instead of using SvelteKit's error handler.
        See comment in +layout.ts
    -->
    {#if isRepoNotFoundErrorLike(data.resolvedRevisionOrError)}
        <RepoNotFoundError repoName={displayRepoName(data.repoName)} {viewerCanAdminister} />
    {:else}
        Something went wrong
    {/if}
{:else}
    <Header>
        <div class="header">
            <nav>
                <h2><a href="/{repo}"><Icon svgPath={mdiSourceRepository} inline /> {displayRepoName(data.repoName)}</a></h2>
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
            {#if data.resolvedRevision}
                <RevisionSelector
                    repoID={data.resolvedRevision.repo.id}
                    revision={data.revision ?? ''}
                    resolvedRevision={data.resolvedRevision}
                    createURL={createRevisionURL}
                />
            {/if}
        </div>
    </Header>
    <slot />
{/if}

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        flex: 1;
    }

    h2 {
        margin: 0;
        margin-right: 1rem;
    }

    nav {
        display: flex;
        align-items: center;
        flex: 1;

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
                background-color: var(--color-bg-2);
            }
        }
    }

    nav {
        color: var(--body-color);
    }
</style>
