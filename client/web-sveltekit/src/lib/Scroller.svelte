<script lang="ts" context="module">
    export interface Capture {
        scroll: number
    }
</script>

<script lang="ts">
    import { afterUpdate, createEventDispatcher } from 'svelte'

    export let margin: number
    export let viewport: HTMLElement | undefined = undefined
    export let scroller: HTMLElement | undefined = undefined

    export function capture(): Capture {
        return { scroll: scroller?.scrollTop || 0 }
    }

    export function restore(data?: Capture) {
        if (!data) return
        // The actual content of the scroller might not be available yet when `restore` is called,
        // e.g. when the data is fetched asynchronously. In that case, we retry a few times.
        let maxTries = 10
        window.requestAnimationFrame(function syncScroll() {
            if (scroller && scroller.scrollTop !== data.scroll) {
                scroller.scrollTop = data.scroll
                if (maxTries > 0) {
                    maxTries -= 1
                    window.requestAnimationFrame(syncScroll)
                }
            }
        })
    }

    const dispatch = createEventDispatcher<{ more: void }>()

    function handleScroll() {
        if (scroller && viewport) {
            const remaining = scroller.scrollHeight - (scroller.scrollTop + (viewport?.clientHeight ?? 0))

            if (remaining < margin) {
                dispatch('more')
            }
        }
    }

    afterUpdate(() => {
        // This premptively triggers a 'more' event when the scrollable content is smaller than than
        // scroller. Without this, the 'more' event would not be triggered because there is nothing
        // to scroll.
        if (scroller && scroller.scrollHeight <= scroller.clientHeight) {
            dispatch('more')
        }
    })
</script>

<div class="viewport" bind:this={viewport} data-viewport>
    <div class="scroller" bind:this={scroller} on:scroll={handleScroll} data-scroller>
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
        overscroll-behavior-y: contain;
    }
</style>
