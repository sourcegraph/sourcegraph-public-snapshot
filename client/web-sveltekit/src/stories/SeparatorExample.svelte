<script lang="ts">
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'

    const currentPosition = getSeparatorPosition('separator-example', 0.5)
    $: width = `${$currentPosition * 100}%`
</script>

<section>
    <div class="left match-highlight" style:min-width={width} style:max-width={width}>Left content</div>
    <Separator {currentPosition} />
    <div class="right">Right content</div>
</section>

<style lang="scss">
    section {
        display: flex;
        height: 90vh;
    }

    div {
        padding: 1rem;
        display: flex;
        align-items: center;
        justify-content: center;
        background-clip: content-box;
    }

    .left {
        background-color: var(--color-bg-2);
    }

    .right {
        flex: 1;
        background-color: var(--color-bg-2);
    }
</style>
