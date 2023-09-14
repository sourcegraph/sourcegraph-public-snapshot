<script lang="ts" context="module">
    export interface Capture {
        scroll: number
    }
</script>

<script lang="ts">
    import { createEventDispatcher } from 'svelte'

    export let margin: number

    export function capture(): Capture {
        return { scroll: scroller.scrollTop }
    }

    export function restore(data: Capture) {
        scroller.scrollTop = data.scroll
    }

    const dispatch = createEventDispatcher<{ more: void }>()

    let viewport: HTMLElement
    let scroller: HTMLElement

    function handleScroll() {
        const remaining = scroller.scrollHeight - (scroller.scrollTop + viewport.clientHeight)

        if (remaining < margin) {
            dispatch('more')
        }
    }
</script>

<div class="viewport" bind:this={viewport}>
    <div class="scroller" bind:this={scroller} on:scroll={handleScroll}>
        <slot />
    </div>
</div>

<style lang="scss">
    .viewport {
        width: 100%;
        height: 100%;
        overflow: hidden;
    }

    .scroller {
        width: 100%;
        height: 100%;
        overflow-y: auto;
    }
</style>
