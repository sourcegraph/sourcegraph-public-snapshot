<script lang="ts" context="module">
    import { treePageSidebarOpen } from "./stores"

    export let sidebarClass: Action<HTMLElement, string> = (node, cls) => {
        const unsubscribe = treePageSidebarOpen.subscribe(isOpen => {
            node.classList.toggle(cls, isOpen)
        })

        return {
            destroy() {
                unsubscribe()
            }
        }
    }
</script>

<script lang="ts">
    import Icon from "$lib/Icon.svelte"
    import Button from "$lib/wildcard/Button.svelte"
    import { mdiChevronDoubleLeft, mdiChevronDoubleRight } from "@mdi/js"
    import type { Action } from "svelte/action"

    export let showWhen: 'open'|'closed'|'always' = 'always'

    let show: boolean;
    $: switch (showWhen) {
        case 'always':
            show = true
            break;
        case 'closed':
            show = !$treePageSidebarOpen
            break;
        case 'open':
            show = $treePageSidebarOpen
            break;
    }

</script>

{#if show}
    <Button variant="secondary" outline size="sm" on:click={() => treePageSidebarOpen.update(open => !open)}>
        <Icon svgPath={$treePageSidebarOpen ? mdiChevronDoubleLeft : mdiChevronDoubleRight} inline />
    </Button>
{/if}
