<script lang="ts">
    import { formatDistanceToNow } from 'date-fns'

    import Avatar from '$lib/Avatar.svelte'

    import type { LastCommitFragment } from './LastCommit.gql'

    export let latestCommit: LastCommitFragment

    $: user = latestCommit.author.person
    $: canonicalURL = latestCommit.canonicalURL
    $: commitMessage = latestCommit.subject
    $: commitDate = formatDistanceToNow(latestCommit.author.date, { addSuffix: false })
</script>

<div class="latest-commit">
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
        margin-right: 0.4rem;
        color: var(--text-muted);
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }

    .display-name {
        margin-left: 0.4rem;
    }

    .latest-commit {
        display: flex;
        flex-flow: row nowrap;
        padding-right: 0.5rem;
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
        margin-right: 0.6rem;
    }
</style>
