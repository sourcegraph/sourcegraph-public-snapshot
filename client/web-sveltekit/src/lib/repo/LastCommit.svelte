<script lang="ts">
    import { formatDistanceToNow } from 'date-fns'

    import Avatar from '$lib/Avatar.svelte'

    import type { LastCommitFragment } from './LastCommit.gql'

    export let lastCommit: LastCommitFragment

    $: user = lastCommit.author.person
    $: canonicalURL = lastCommit.canonicalURL
    $: commitMessage = lastCommit.subject
    $: commitDate = formatDistanceToNow(lastCommit.author.date, { addSuffix: true })
</script>

<div class="last-commit">
    <div class="avatar">
        <Avatar avatar={user} />
    </div>
    <div class="display-name">
        <span class="label">{user.displayName || user.name}</span>
    </div>
    <div class="commit-message">
        <a href={canonicalURL}>
            {commitMessage}
        </a>
    </div>
    <div class="commit-date">
        {commitDate}
    </div>
</div>

<style lang="scss">
    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: space-between;
        margin-right: 0.5rem;
        white-space: nowrap;
        max-width: 400px;
        font-size: var(--font-size-small);
    }

    .avatar {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        margin-right: 0.5rem;
        --avatar-size: 1.5rem;
    }

    .display-name {
        margin-right: 0.75rem;
        color: var(--text-body);
    }

    .commit-message {
        align-items: center;
        color: var(--text-muted);
        margin-right: 0.5rem;
        max-width: 240px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .commit-date {
        color: var(--text-muted);
    }
</style>
