<script lang="ts">
    export let path: string
</script>

<svg class="icon-inline" viewBox="0 0 24 24">
    <path d={path} />
</svg>

<style lang="scss">
    svg {
        &.icon-inline {
            width: var(--icon-inline-size);
            height: var(--icon-inline-size);
            fill: var(--icon-color);
            vertical-align: text-bottom;
        }
    }
</style>
