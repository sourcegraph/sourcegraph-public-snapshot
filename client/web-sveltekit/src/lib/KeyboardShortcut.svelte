<!-- @component KeyboardShortcut
A component to display the keyboard shortcuts for the application.
-->
<script lang="ts">
    import { isMacPlatform } from '$lib/common'
    import { formatShortcutParts, type Keys } from '$lib/Hotkey'

    export let shortcut: Keys
    export let inline: boolean = false

    const separator = isMacPlatform() ? '' : '+'

    $: parts = (() => {
        const result: string[] = []
        let parts = formatShortcutParts(shortcut)
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
        all: unset;
        display: inline-block;
        &.inline {
            display: inline;
        }

        box-sizing: border-box;
        line-height: 1;
        $height: (20 / 16) * 1em;
        $verticalPadding: $height * 0.1875;
        padding: $verticalPadding ($verticalPadding * 1.5);
        border-radius: 0.375em;
        font-size: $height - $verticalPadding * 2;
        font-family: var(--font-family-base);

        background-color: var(--secondary-4);
        color: var(--text-muted);

        // When inside a selected container, show the selected variant
        :global([aria-selected='true']) & {
            color: white;
            background-color: var(--primary);
        }
    }
</style>
