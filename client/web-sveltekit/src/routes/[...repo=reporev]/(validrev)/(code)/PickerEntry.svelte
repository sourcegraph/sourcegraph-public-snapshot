<script lang="ts">
    import Avatar from "$lib/Avatar.svelte"
    import Icon from "$lib/Icon.svelte"
    import Timestamp from "$lib/Timestamp.svelte"
    import type { GitCommit } from "$lib/graphql-types"

    export let iconPath: string
    export let label: string
    export let commit: GitCommit|null
</script>

<span class="title">
    <slot name="title">
        <Icon svgPath={iconPath} inline />
        <span>{label}</span>
    </slot>
</span>
<span class="author">
    {#if commit}
        <Avatar avatar={commit.author.person} />
        {commit.author.person.displayName}
    {/if}
</span>
<span class="timestamp">
    {#if commit}
        <Timestamp date={commit.author.date} strict />
    {/if}
</span>

<style lang="scss">
    .title,
    .author,
    .timestamp {
        display: flex;
        gap: 0.25rem;
        align-items: center;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .title {
        padding-left: 0.75rem;
        padding-right: 0.5rem;

        // Branch icon
        :global(svg) {
            flex-shrink: 0;
        }

        // Branch name badge
        :global(span) {
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }
</style>
