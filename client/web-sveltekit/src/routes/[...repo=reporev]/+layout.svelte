<script lang="ts">
    import { mdiAccount, mdiCodeTags, mdiSourceBranch, mdiSourceCommit, mdiSourceRepository, mdiTag } from '@mdi/js'

    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import { displayRepoName } from '$lib/shared'

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

    $: ({ repo } = $page.params)

    $: repoName = displayRepoName(repo.split('@')[0])
</script>

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
</div>
<slot />

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem;
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

    nav {
        color: var(--body-color);
    }
</style>
