<!-- @component KeyboardShortcut
A component to display the keyboard shortcuts for the application.
-->
<script lang="ts">
    import { isMacPlatform } from '$lib/common'
    import { formatShortcutParts, type Keys } from '$lib/Hotkey'

    export let shorcut: Keys
    export let inline: boolean = false

    const separator = isMacPlatform() ? '' : '+'

    $: parts = (() => {
        const result: string[] = []
        let parts = formatShortcutParts(shorcut)
        for (let i = 0; i < parts.length; i++) {
            if (i > 0) {
                result.push(separator)
            }
            result.push(parts[i])
        }
        return result
    })()
</script>

<kbd class:inline>
    {#each parts as part}
        <span>{part}</span>
    {/each}
</kbd>

<style lang="scss">
    kbd {
        display: inline-flex;
        align-items: center;
        gap: 0.125rem;
    }

    .inline {
        margin: 0;
        padding: 0;
        border: none;
        box-shadow: none;
        background-color: transparent;
    }
</style>
