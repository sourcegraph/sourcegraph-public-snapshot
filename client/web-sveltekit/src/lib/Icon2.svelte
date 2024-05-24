<!--
    @component
    Creates an SVG icon. You can overwrite the color by using the --color
    style directive:

    <Icon svgPath={...} --color="other color" />

    Otherwise the current text color is used.
-->
<script lang="ts">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'
    import IconClose from 'virtual:icons/lucide/x'

    const icons = {
        close: IconClose,
    }

    type $$Props = SvelteHTMLElements['svg'] & {
        icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>> | 'close'
        inline?: boolean
    }

    export let icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>> | 'close'
    export let inline: boolean = false

    $: resolvedIcon = typeof icon === 'string' ? icons[icon] : icon
</script>

<svelte:component this={resolvedIcon} class="icon {inline ? 'icon-inline' : ''}" />

<style lang="scss">
    $iconSize: var(--icon-size, 1.5rem);
    $iconInlineSize: var(--icon-inline-size, #{(16 / 14)}em);

    :global(svg.icon) {
        width: $iconSize;
        height: $iconSize;

        color: var(--icon-fill-color, var(--color, inherit));
        fill: currentColor;
    }

    :global(svg.icon-inline) {
        width: $iconInlineSize;
        height: $iconInlineSize;
        vertical-align: text-bottom;
    }
</style>
