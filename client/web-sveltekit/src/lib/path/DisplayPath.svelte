<!--
    @component
    DisplayPath is a path that is formatted for display.

    Some features it provides:
    - Styleable slashes (target data-slash) and path items (target data-path-item)
    - Styleable spacing (target gap in data-path-container)
    - An optional copy button which copies the full path (target data-copy-button)
    - An optional icon before the last path element
    - An optional callback to linkify a path item
    - Zero additional whitespace in the DOM
        - This means document.querySelector('[data-path-container]').textContent should always exactly equal the path
    - Selecting the path and copying it manually still works, even with inline icons (which would normally add a space)

    This component is designed to be styled by targeting data- attributes rather than by using slots
    because it is disturbingly easy to break the whitespace and path copying guarantees this component
    provides by introducing whitespace or inline block elements in a slot. That is also why this is
    a somwhat "full-featured" component rather than being designed more around composition.
-->
<script lang="ts">
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    /**
     * The path to be formatted
     */
    export let path: string
    /**
     * Whether to show a "copy path" button. Can be styled by targeting `data-copy-button`
     */
    export let showCopyButton = false
    /**
     * A callback to generate an href for the given path. If unset, path items
     * will not be linkified. Will be called with the full path prefix for a
     * path item. For example, for the path `tmp/test`, it will be called with
     * `tmp` and `tmp/test`.
     *
     * For most cases, use the `pathHrefFactory` helper to create this callback.
     */
    export let pathHref: ((path: string) => string) | undefined = undefined

    $: parts = path.split('/').map((part, index, allParts) => ({ part, path: allParts.slice(0, index + 1).join('/') }))
</script>

<!--
    NOTE: all the weird comments in here are being very careful to not
    introduce any additional whitespace in the path container since that would
    make the path invalid when copied from a selection
-->
<span data-path-container>
    <slot name="prefix" /><!--
    -->{#each parts as { part, path }, index}<!--
        -->{@const last =
            index === parts.length - 1}<!--
        -->{#if index > 0}<!--
            --><span data-slash>/</span
            ><!--
        -->{/if}<!--
        Wrap the anchor element with a span because otherwise it adds
        spaces around it when copied
        --><span
            class:last
            data-path-item
            >{#if pathHref}<a href={pathHref(path)}>{part}</a>{:else}{part}{/if}<!--
            --></span
        ><!--
        -->{#if last}<!--
            --><span data-file-icon aria-hidden="true"
                ><slot name="file-icon" /></span
            ><!--
        -->{/if}<!--
    -->{/each}<!--
    We include the copy button in this component because we want it to wrap along with the path
    elements. Otherwise, an invisible button might wrap to its own line, which looks weird.
    -->{#if showCopyButton}<!--
        --><span
            data-copy-button><CopyButton value={path} label="Copy path to clipboard" /></span
        ><!--
    -->{/if}
</span>

<style lang="scss">
    [data-path-container] {
        display: inline-flex;
        align-items: center;
        gap: 0.125em;

        white-space: pre-wrap;

        font-weight: 400;
        font-size: var(--code-font-size);
        font-family: var(--code-font-family);

        // Global so data-slash can be slotted in with the prefix
        // and styled consistently.
        :global([data-slash]) {
            color: var(--text-disabled);
            display: inline;
        }
    }

    [data-path-item] {
        display: inline;
        white-space: nowrap;
        color: var(--text-body);
        & > a {
            color: inherit;
        }
    }

    // HACK: The file icon is placed after the file name in the DOM so it
    // doesn't add any spaces in the file path when copied. This visually
    // reorders the last path element after the file icon.
    .last {
        order: 1;
    }

    [data-file-icon] {
        user-select: none; // Avoids a trailing space on select + copy
        &:empty {
            display: none;
        }
    }

    [data-copy-button] {
        order: 1;
        margin-left: 0.5rem;
        user-select: none; // Avoids a trailing space on select + copy
    }
</style>
