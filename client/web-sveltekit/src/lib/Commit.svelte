<svelte:options immutable />

<script lang="ts">
    import { mdiDotsHorizontal } from '@mdi/js'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import type { Commit } from './Commit.gql'

    export let commit: Commit
    export let alwaysExpanded: boolean = false

    function getCommitter({ committer }: Commit): NonNullable<Commit['committer']> | null {
        if (!committer) {
            return null
        }
        // Do not show if committer is GitHub (e.g. squash merge)
        if (committer.person.name === 'GitHub' && committer.person.email === 'noreply@github.com') {
            return null
        }
        return committer
    }

    $: author = commit.author
    $: committer = getCommitter(commit) ?? author
    $: committerIsAuthor = committer.person.email === author.person.email
    $: commitDate = new Date(committer.date)
    $: authorAvatarTooltip = author.person.name + (committer ? ' (author)' : '')
    let expanded = alwaysExpanded
</script>

<div class="root">
    <div class="avatar">
        <Tooltip tooltip={authorAvatarTooltip}>
            <Avatar avatar={author.person} />
        </Tooltip>
    </div>
    {#if !committerIsAuthor}
        <div class="avatar">
            <Tooltip tooltip="{committer.person.name} (committer)">
                <Avatar avatar={committer.person} />
            </Tooltip>
        </div>
    {/if}
    <div class="info">
        <span class="title">
            <a class="subject" href={commit.canonicalURL}>{commit.subject}</a>
            {#if !alwaysExpanded && commit.body}
                <button type="button" on:click={() => (expanded = !expanded)}>
                    <Icon svgPath={mdiDotsHorizontal} inline />
                </button>
            {/if}
        </span>
        <span>
            {#if !committerIsAuthor}authored by <strong>{author.person.name}</strong> and{/if}
            committed by <strong>{committer.person.name}</strong>
            <Timestamp date={commitDate} />
        </span>
        {#if expanded && commit.body}
            <pre>{commit.body}</pre>
        {/if}
    </div>
</div>

<style lang="scss">
    .root {
        display: flex;
        gap: 1rem;
    }

    .info {
        display: flex;
        flex-direction: column;
        flex: 1;
        min-width: 0;
    }

    .title {
        display: flex;
        gap: 0.5rem;

        .subject {
            font-weight: 600;
            flex: 0 1 auto;
            color: var(--body-color);
            min-width: 0;

            @media (--sm-breakpoint-up) {
                overflow: hidden;
                white-space: nowrap;
                text-overflow: ellipsis;
            }
        }
    }

    .avatar {
        flex: 0 0 auto;
        display: flex;
        width: 2.75rem;
        height: 2.75rem;
        font-size: 1.5rem;
    }

    span {
        color: var(--text-muted);
    }

    button {
        color: var(--body-color);
        border: 1px solid var(--secondary);
        cursor: pointer;

        @media (--xs-breakpoint-down) {
            align-self: flex-start;
        }
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
</style>
