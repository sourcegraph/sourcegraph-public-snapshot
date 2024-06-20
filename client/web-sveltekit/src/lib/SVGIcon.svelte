<!--
    @deprecated Use Icon.svelte instead.
    @component
    Creates an SVG icon. You can overwrite the color by using the --color
    style directive:

    <SVGIcon svgPath={...} --color="other color" />

    Otherwise the current text color is used.
-->
<script lang="ts">
    import type { SVGAttributes } from 'svelte/elements'

    type $$Props = SVGAttributes<SVGElement> & {
        svgPath: string
        inline?: boolean
    }

    export let svgPath: string
    export let inline: boolean = false
</script>

<svg class:icon-inline={inline} viewBox="0 0 24 24" data-icon {...$$restProps}>
    <path d={svgPath} />
</svg>

<style lang="scss">
    $iconSize: var(--icon-size, 1.5rem);
    $iconInlineSize: var(--icon-inline-size, #{(16 / 14)}em);

    svg {
        width: $iconSize;
        height: $iconSize;
        &.icon-inline {
            width: $iconInlineSize;
            height: $iconInlineSize;
            vertical-align: text-bottom;
        }
        // Note that --icon-color is defined globally
        color: var(--icon-color, inherit);
        fill: currentColor;
    }
</style>
