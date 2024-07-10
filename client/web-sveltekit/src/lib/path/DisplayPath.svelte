<script lang="ts">
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    export let path: string
    export let showCopyButton = false
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
            >{#if pathHref}<a href={pathHref(path)}>{part}</a>{:else}{part}{/if}<!--component
            --></span
        ><!--
        -->{#if last}<!--
            --><span data-file-icon aria-hidden="true"
                ><slot name="file-icon" /></span
            ><!--
        -->{/if}<!--
    -->{/each}<!--
    -->{#if showCopyButton}<!--
        --><span data-copy-button
            ><CopyButton value={path} label="Copy path to clipboard" /></span
        ><!--
    -->{/if}
</span>

<style lang="scss">
    [data-path-container] {
        display: inline-flex;
        align-items: center;
        gap: 0.125em;

        font-weight: 400;
        font-size: var(--code-font-size);
        font-family: var(--code-font-family);
    }

    // Global so a data-slash can be slotted in in the prefix
    [data-slash] {
        color: var(--text-disabled);
        display: inline;
    }

    [data-path-item] {
        display: inline;
        white-space: nowrap;
        color: var(--text-body);
        & > a {
            color: inherit;
        }

        // HACK: The file icon is placed after the file name
        // in the DOM so it doesn't add any spaces in the file
        // path when copied. This visually reorders the last
        // path element after the file icon.
        &.last {
            order: 1;
        }
    }

    [data-file-icon] {
        user-select: none; // Avoids a trailing space on select + copy
        &:empty {
            display: none;
        }
    }

    [data-copy-button] {
        margin-left: 0.5rem;
        user-select: none; // Avoids a trailing space on select + copy
        order: 1;
    }
</style>
