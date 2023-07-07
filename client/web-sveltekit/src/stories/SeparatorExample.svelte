<script lang="ts">
    import '../routes/styles.scss'

    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'

    const currentPosition = getSeparatorPosition('separator-example', 0.5)
    $: width = `${$currentPosition * 100}%`
</script>

<section>
    <div class="left" style:min-width={width} style:max-width={width}>Left content</div>
    <Separator {currentPosition} />
    <div class="right">Right content</div>
</section>

<style lang="scss">
    section {
        // Doesn't look like border-color is set for whatever reason, making the separator invisible
        --border-color: black;
        display: flex;
        height: 90vh;
    }

    div {
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .left {
        background-color: lightblue;
    }

    .right {
        flex: 1;
        background-color: lightgray;
    }
</style>
