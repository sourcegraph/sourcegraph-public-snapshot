<script lang="ts">
    import { mdiDotsHorizontal } from '@mdi/js'

    import type { GitCommitFields } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import Timestamp from './Timestamp.svelte'

    export let commit: GitCommitFields
    export let alwaysExpanded: boolean = false

    $: commitDate = new Date(commit.committer ? commit.committer.date : commit.author.date)
    let expanded = alwaysExpanded
</script>

<div class="root">
    <div class="avatar">
        <UserAvatar user={commit.author.person} />
    </div>
    {#if commit.committer}
        <div class="avatar">
            <UserAvatar user={commit.committer.person} />
        </div>
    {/if}
    <div class="info">
        <span class="d-flex">
            <a class="subject" href={commit.url}>{commit.subject}</a>
            {#if !alwaysExpanded}
                <button type="button" on:click={() => (expanded = !expanded)}>
                    <Icon svgPath={mdiDotsHorizontal} inline />
                </button>
            {/if}
        </span>
        <span>committed by <strong>{commit.author.person.name}</strong> <Timestamp date={commitDate} /></span>
        {#if expanded}
            <pre>{commit.body}</pre>
        {/if}
    </div>
    {#if !alwaysExpanded}
        <div class="buttons">
            <a href={commit.url}>{commit.abbreviatedOID}</a>
        </div>
    {/if}
</div>

<style lang="scss">
    .root {
        display: flex;
    }

    .info {
        display: flex;
        flex-direction: column;
        margin: 0 0.5rem;
        flex: 1;
        min-width: 0;
    }

    .subject {
        font-weight: 600;
        flex: 0 1 auto;
        padding-right: 0.5rem;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
        color: var(--body-color);
        min-width: 0;
    }

    .avatar {
        flex: 0 0 auto;
        display: flex;
        width: 2.75rem;
        height: 2.75rem;
        margin-right: 0.5rem;
        font-size: 1.5rem;
    }

    span {
        color: var(--text-muted);
    }

    button {
        color: var(--body-color);
        border: 1px solid var(--secondary);
        cursor: pointer;
    }
    pre {
        margin-top: 0.5rem;
        margin-bottom: 1.5rem;
        font-size: 0.75rem;
        overflow: visible;
        max-width: 100%;
        word-wrap: break-word;
        white-space: pre-wrap;
    }

    .buttons {
        align-self: center;

        a {
            display: inline-block;
            padding: 0.125rem;
            font-family: var(--code-font-family);
            font-size: 0.75rem;
        }
    }
</style>
