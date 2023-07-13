<script lang="ts">
    import { onMount } from 'svelte'

    import { afterNavigate, disableScrollHandling } from '$app/navigation'
    import { page } from '$app/stores'
    import FileTree from '$lib/repo/FileTree.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import { sidebarOpen } from '$lib/repo/stores'
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'
    import { scrollAll } from '$lib/stores'
    import { asStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    function last<T>(arr: T[]): T {
        return arr[arr.length - 1]
    }

    $: treeOrError = asStore(data.treeEntries.deferred)

    const sidebarSize = getSeparatorPosition('repo-sidebar', 0.2)
    $: sidebarWidth = `max(200px, ${$sidebarSize * 100}%)`

    onMount(() => {
        // We want the whole page to be scrollable and hide page and repo navigation
        scrollAll.set(true)
        return () => scrollAll.set(false)
    })

    afterNavigate(() => {
        // Prevents SvelteKit from resetting the scroll position to the top
        disableScrollHandling()
    })
</script>

<section>
    <div class="sidebar" class:open={$sidebarOpen} style:min-width={sidebarWidth} style:max-width={sidebarWidth}>
        {#if !$treeOrError.loading && $treeOrError.data}
            <FileTree
                activeEntry={$page.params.path ? last($page.params.path.split('/')) : ''}
                treeOrError={$treeOrError.data}
            >
                <h3 slot="title">
                    <SidebarToggleButton />&nbsp; Files
                </h3>
            </FileTree>
        {/if}
    </div>
    {#if $sidebarOpen}
        <Separator currentPosition={sidebarSize} />
    {/if}
    <div class="content">
        <slot />
    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        flex: 1;
        flex-shrink: 0;
        background-color: var(--code-bg);
        min-height: 100vh;
    }

    .sidebar {
        &.open {
            display: flex;
            flex-direction: column;
        }
        display: none;
        overflow: hidden;
        background-color: var(--body-bg);
        padding: 0.5rem;
        padding-bottom: 0;
        position: sticky;
        top: 0px;
        max-height: 100vh;
    }

    .content {
        flex: 1;
        display: flex;
        flex-direction: column;
        min-width: 0;
    }

    h3 {
        display: flex;
        align-items: center;
        margin-bottom: 0.5rem;
    }
</style>
