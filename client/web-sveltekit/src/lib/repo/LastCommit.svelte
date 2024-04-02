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
    <div class="user-info">
        <Avatar avatar={user} />
        <div class="display-name">
            <small>
                {user.name}
            </small>
        </div>
    </div>

    <div class="commit-message">
        <a href={canonicalURL}>
            <small>
                {commitMessage}
            </small>
        </a>
    </div>

    <div class="commit-date">
        <small>
            {commitDate}
        </small>
    </div>
</div>

<style lang="scss">
    .commit-date {
        color: var(--text-muted);
    }

    .commit-message {
        margin-right: 0.5rem;
        color: var(--text-muted);
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }

    .display-name {
        margin-left: 0.5rem;
    }

    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        max-width: 350px;
        align-items: center;
        white-space: nowrap;
        justify-content: space-between;
        margin-right: 0.5rem;
    }

    .user-info {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        margin-right: 0.5rem;
    }
</style>
