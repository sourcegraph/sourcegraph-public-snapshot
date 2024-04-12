<script lang="ts">
    import type { Avatar_Person, Avatar_User } from '$lib/Avatar.gql'
    import Avatar from '$lib/Avatar.svelte'
    import type { BlameHunk } from '$lib/web'

    export let hunk: BlameHunk

    // These are in the React implementation, but not yet used in the Svlete implementation.
    // Leaving them commented for reference if/when we implement the popover.
    //
    // export let line: number
    // export let onSelect: (line: number) => void
    // export let onDeselect: (line: number) => void
    // export let externalURLs: BlameHunkData['externalURLs']

    $: info = hunk.displayInfo

    function getAvatar(author: BlameHunk['author']): Avatar_Person | Avatar_User {
        // Avatar expects GraphQL types, but since these come from the
        // blame view, our types don't quite match. Massage them a bit
        // to get them in the right shape.
        if (author.person.user) {
            return { __typename: 'User', ...author.person.user }
        } else {
            return {
                __typename: 'Person',
                displayName: author.person.displayName,
                name: author.person.displayName,
                avatarURL: author.person.avatarURL,
            }
        }
    }
</script>

<div class="root">
    <span class="date">{info.dateString}</span>
    <Avatar avatar={getAvatar(hunk.author)} />
    <a href={info.linkURL} target="_blank" rel="noreferrer noopener">
        {`${info.displayName}${info.username}`.split(' ')[0]}
        {' • '}
        <span class="message">{info.message}</span>
    </a>
</div>

<style lang="scss">
    .root {
        font-family: var(--font-family-base);
        color: var(--text-muted);
        display: flex;
        align-items: center;
        padding: 0 0.5em;
        gap: 0.5em;

        .date {
            min-width: 80px;
        }

        a {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: inherit;
        }
    }
</style>
