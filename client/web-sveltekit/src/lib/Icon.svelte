<script lang="ts">
    export let ariaLabel: string = ''
    export let svgPath: string
    export let inline: boolean = false
    let className = ''
    export { className as class }
</script>

<svg class={className} class:icon-inline={inline} height="24" width="24" viewBox="0 0 24 24" aria-label={ariaLabel}>
    <path d={svgPath} />
</svg>

<style lang="scss">
    svg {
        color: var(--color, inherit);
        fill: currentColor;
    }
</style>
