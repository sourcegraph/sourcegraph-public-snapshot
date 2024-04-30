<!--
    @component
    Creates an SVG icon. You can overwrite the color by using the --color
    style directive:

    <Icon svgPath={...} --color="other color" />

    Otherwise the current text color is used.
-->
<script lang="ts">
    import type { SVGAttributes } from 'svelte/elements'

    interface $$Props extends SVGAttributes<SVGElement> {
        svgPath: string
        inline?: boolean
        size?: number
    }

    export let svgPath: string
    export let inline: boolean = false
</script>

<svg class:icon-inline={inline} viewBox="0 0 24 24" data-icon {...$$restProps}>
    <path d={svgPath} />
</svg>

<style lang="scss">
    :root {
        --icon-size: 1.5rem;
        --icon-inline-size: #{(16 / 14)}em;
    }

    svg {
        width: var(--icon-size);
        height: var(--icon-size);
        &.icon-inline {
            width: var(--icon-inline-size);
            height: var(--icon-inline-size);
        }
        color: var(--icon-fill-color, var(--color, inherit));
        fill: currentColor;
    }
</style>
