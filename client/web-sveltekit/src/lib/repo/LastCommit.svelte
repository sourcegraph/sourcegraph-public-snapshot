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
    <!-- Wrapper diff necessary to prevent avatar from being squished -->
    <div>
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
        --avatar-size: 1.5rem;

        display: flex;
        align-items: center;
        gap: 0.5rem;
        white-space: nowrap;
        font-size: var(--font-size-small);
    }

    .display-name {
        margin-right: 0.25rem;
        color: var(--text-body);
    }

    .commit-message {
        color: var(--text-muted);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .commit-date {
        color: var(--text-muted);
    }
</style>
