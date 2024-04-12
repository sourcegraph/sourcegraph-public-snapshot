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
        <small>{user.name}</small>
    </div>
    <div class="commit-message">
        <a href={canonicalURL}>
            <small>{commitMessage}</small>
        </a>
    </div>
    <div class="commit-date">
        <small>{commitDate}</small>
    </div>
</div>

<style lang="scss">
    .avatar {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        margin-right: 0.25rem;
    }

    .commit-date {
        color: var(--text-muted);
    }

    .commit-message {
        align-items: center;
        color: var(--text-muted);
        margin-right: 0.5rem;
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .display-name {
        margin-right: 0.5rem;
    }

    .last-commit {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        margin-right: 0.5rem;
        white-space: nowrap;
        max-width: 400px;
    }
</style>
