<script lang="ts">
    import { page } from '$app/stores'
    import Shimmer from '$lib/Shimmer.svelte'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import type { GitCommit } from '$lib/graphql-operations'
    import { getRelativeTime } from '$lib/relativeTime'
    import { navFromPath } from '../utils'
    import SidebarToggleButton, { sidebarClass } from './SidebarToggleButton.svelte'

    export let commit: Promise<GitCommit|null>|undefined = undefined

    $: breadcrumbs = navFromPath($page.params.path, $page.params.repo)
</script>

<div class="header">
    <div class="toggle-wrapper" use:sidebarClass={"open"}>
        <SidebarToggleButton />
    </div>
    <h2>
        <span class="icon">
            <slot name="icon"/>&nbsp;
        </span>
        <span>
        {#each breadcrumbs as [name, path], index}
                {#if index > 0}
                 /
                {/if}
            <span class:last={index === breadcrumbs.length - 1}>
                {#if path}
                    <a href={path}>{name}</a>
                {:else}
                    {name}
                {/if}
            </span>
        {/each}
        </span>
    </h2>
    <div class="actions">
        <slot name="actions"/>
    </div>
</div>
{#if commit}
    <div class="commit">
        {#await commit}
            <Shimmer --height="24px" />
        {:then commit}
            {#if commit}
                <span class="user">
                    <UserAvatar user={commit.author.person} />&nbsp;
                    <span class="text-muted">{commit.author.person.displayName}</span>&nbsp;
                </span>
                <a href={commit.url}>{commit.subject}</a>
                <span>
                    <a href={commit.url}>{commit.abbreviatedOID}</a>
                    &nbsp;&middot;&nbsp;
                    <span>{getRelativeTime(new Date(commit.author.date))}</span>
                </span>
            {/if}
        {/await}
    </div>
{/if}

<style lang="scss">
.header {
    display: flex;
    align-items: center;
}

.toggle-wrapper {
    display: initial;
    margin-right: 0.1rem;

    &:global(.open) {
        display: none;
    }
}

h2 {
    display: flex;
    align-items: center;
    font-weight: normal;
    font-size: 1rem;
    margin: 0;

    .last {
        font-weight: bold;
    }

    .icon {
        flex-shrink: 0;

    }
}

.actions {
    margin-left: auto;
}

.commit {
    border-top: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 0.5rem;
    padding-top: 0.5rem;
}

.user {
    flex-shrink: 0;
}

a {
    flex: 1;
}

</style>
