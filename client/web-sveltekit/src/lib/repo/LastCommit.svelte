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
    <span class="display-name">{user.displayName || user.name}</span>
    <a href={canonicalURL}>
        {commitMessage}
    </a>
    <span class="timestamp"><Timestamp date={lastCommit.author.date} /></span>
</div>

<style lang="scss">
    .display-name {
        padding-right: 0.25rem;
    }
    .last-commit {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        white-space: nowrap;
        font-size: var(--font-size-small);
    }

    a {
        flex: 1;
        color: var(--text-muted);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .timestamp {
        color: var(--text-muted);
    }
</style>
