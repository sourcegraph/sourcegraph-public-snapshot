<script lang="ts">
    import Avatar from '$lib/Avatar.svelte'
    import Icon2 from '$lib/Icon2.svelte'
    import type { IconPath } from '$lib/Icon2.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { Badge } from '$lib/wildcard'

    import type { RepositoryGitRevAuthor } from './RepositoryRevPicker.gql'

    export let iconPath: IconPath
    export let label: string
    export let author: RepositoryGitRevAuthor['author'] | null | undefined
</script>

<span class="title">
    <slot name="title">
        <Icon2 icon={iconPath} inline />
        <Badge variant="link">{label}</Badge>
    </slot>
</span>
<span class="author">
    {#if author}
        <Avatar avatar={author.person} />
        <span class="author-name">{author.person.displayName}</span>
    {/if}
</span>
<span class="timestamp">
    {#if author}
        <Timestamp date={author.date} strict />
    {/if}
</span>

<style lang="scss">
    .title,
    .author,
    .timestamp {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;

        // Prevent avatar image from shrinking
        :global([data-avatar]) {
            --avatar-size: 1.5rem;

            flex-shrink: 0;
        }

        // Timestamp uses tooltip wrapper element with display:contents
        // override this behavior since we have to overflow text within
        // trigger text
        .author-name,
        :global([data-tooltip-root]) {
            display: block;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .title {
        padding-left: 1rem;

        :global([data-icon]) {
            flex-shrink: 0;
            color: var(--icon-muted);

            :global([data-highlighted]) &,
            :global([data-picker-suggestions-list-item]):hover & {
                color: var(--icon-color);
            }
        }

        // Branch name badge
        :global([data-badge]) {
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .timestamp {
        padding-right: 0.75rem;
        color: var(--text-muted);
    }
</style>
