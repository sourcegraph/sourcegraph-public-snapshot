<script lang="ts">
    import Avatar from '$lib/Avatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'

    import type { LastCommitFragment } from './LastCommit.gql'

    export let lastCommit: LastCommitFragment

    $: user = lastCommit.author.person
    $: canonicalURL = lastCommit.canonicalURL
    $: commitMessage = lastCommit.subject
</script>

<div class="last-commit">
    <Avatar avatar={user} --avatar-size="1.5rem" />
    <span>{user.displayName || user.name}</span>
    <a href={canonicalURL}>
        {commitMessage}
    </a>
    <span class="timestamp"><Timestamp date={lastCommit.author.date} /></span>
</div>

<style lang="scss">
    .last-commit {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        white-space: nowrap;
        font-size: var(--font-size-small);
    }

    a {
        color: var(--text-muted);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .timestamp {
        color: var(--text-muted);
    }
</style>
