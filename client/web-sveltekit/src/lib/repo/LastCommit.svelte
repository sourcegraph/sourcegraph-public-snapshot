<script lang="ts" context="module">
    export interface LastCommitProps {
        id: string
        abbreviatedOID: string
        subject: string
        canonicalURL: string
        author: {
            date: string
            person: {
                avatarURL: string | null
                displayName: string
                name: string
            }
        }
    }
</script>

<script lang="ts">
    import { formatDistanceToNow } from 'date-fns'

    import type { Avatar_User } from '$lib/Avatar.gql'
    import Avatar from '$lib/Avatar.svelte'

    export let latestCommit: LastCommitProps

    let avatar: Avatar_User = {
        __typename: 'User',
        avatarURL: latestCommit.author.person.avatarURL,
        displayName: latestCommit.author.person.displayName,
        username: latestCommit.author.person.name,
    }
</script>

<div class="latest-commit">
    <div class="user-info">
        <Avatar {avatar} />
        <div class="display-name">
            <small>
                {latestCommit.author.person.name}
            </small>
        </div>
    </div>

    <div class="commit-message">
        <a href={latestCommit.canonicalURL}>
            <small>
                {latestCommit.subject}
            </small>
        </a>
    </div>

    <div class="commit-date">
        <small>
            {formatDistanceToNow(latestCommit.author.date, { addSuffix: false })}
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
