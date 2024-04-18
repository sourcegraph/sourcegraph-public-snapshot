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
    export let size: number = 24
</script>

<svg class:icon-inline={inline} height={size} width={size} viewBox="0 0 24 24" data-icon {...$$restProps}>
    <path d={svgPath} />
</svg>

<style lang="scss">
    svg {
        color: var(--color, inherit);
        fill: currentColor;
    }
</style>
