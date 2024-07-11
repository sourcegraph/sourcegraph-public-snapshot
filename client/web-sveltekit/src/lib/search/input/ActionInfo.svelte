<script lang="ts">
    import type { Action } from '$lib/branded'
    import type { Keys } from '$lib/Hotkey'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'

    export let action: Action
    export let shortcut: Keys
</script>

Press <KeyboardShortcut {shortcut} inline /> to
{#if action.type === 'completion'}
    <strong>add</strong> to your query.
{:else if action.type === 'goto'}
    <strong>go to</strong> the suggestion.
{:else if action.type === 'command'}
    <strong>execute</strong> the command.
{/if}
