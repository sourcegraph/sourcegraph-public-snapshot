<!-- @component KeyboardShortcut
A component to display the keyboard shortcuts for the application.
-->
<script lang="ts">
    import { isMacPlatform } from '$lib/common'
    import { formatShortcutParts, type Keys } from '$lib/Hotkey'
    import { isViewportMobile } from './stores'

    export let shortcut: Keys
    export let inline: boolean = false

    const separator = isMacPlatform() ? '' : '+'

    // No need to do this work if we are on a mobile device
    $: parts = $isViewportMobile
        ? []
        : (() => {
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

{#if !$isViewportMobile}
    <kbd class:inline>
        {#each parts as part}
            <span>{part}</span>
        {/each}
    </kbd>
{/if}

<style lang="scss">
    kbd {
        all: unset;
        display: inline-block;
        &.inline {
            display: inline;
        }

        // NOTE: the height of the kdb element is based on the base
        // line height. There is no way to query the line height of the
        // parent, so we assume that the container has the standard
        // line height. In the case the parent has a line height of 1,
        // this will grow the container slightly. This can be overridden
        // by setting `--line-height-base` if overriding it in the parent.
        --height: calc(var(--line-height-base) * 1em);
        --vertical-padding: calc(var(--height) * 0.25);
        padding: var(--vertical-padding) calc(1.5 * var(--vertical-padding));
        font-size: calc(var(--height) - 2 * var(--vertical-padding));
        font-family: var(--font-family-base);
        border-radius: 0.5em;
        line-height: 1;

        background-color: var(--secondary-4);
        color: var(--text-body);

        // When inside a selected container, show the selected variant
        :global([aria-selected='true']) & {
            color: var(--primary);
        }
    }
</style>
