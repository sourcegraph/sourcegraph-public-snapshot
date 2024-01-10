<script lang="ts">
    import type { Action } from '$lib/branded'
    import { shortcutDisplayName } from '$lib/shared'

    export let action: Action
    export let shortcut: string

    $: displayName = shortcutDisplayName(shortcut)
</script>

Press <kbd>{displayName}</kbd> to
{#if action.type === 'completion'}
    <strong>add</strong> to your query
{:else if action.type === 'goto'}
    <strong>go to</strong> the suggestion
{:else if action.type === 'command'}
    <strong>execute</strong> the command
{/if}

<style lang="scss">
    kbd {
        all: unset;
        // This is the code live on S2 but I can't ind it in our code base
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
        display: inline-block;
        line-height: 1.3333333333;
        height: 1.125rem;
        padding: 0 0.25rem;
        margin: 0 0.125rem;
        vertical-align: middle;
        border-radius: 3px;
        color: var(--body-color);
        background-color: var(--color-bg-2);
        box-shadow: inset 0 -2px 0 var(--color-bg-3);
    }
</style>
