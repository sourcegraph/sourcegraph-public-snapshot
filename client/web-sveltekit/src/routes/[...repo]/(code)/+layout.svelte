<script lang="ts">
    import { mdiChevronDoubleLeft, mdiChevronDoubleRight } from '@mdi/js'

    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import FileTree from '$lib/repo/FileTree.svelte'
    import { asStore } from '$lib/utils'
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'

    import type { PageData } from './$types'

    export let data: PageData

    function last<T>(arr: T[]): T {
        return arr[arr.length - 1]
    }

    $: treeOrError = asStore(data.treeEntries.deferred)

    let showSidebar = true
    const sidebarSize = getSeparatorPosition('repo-sidebar', 0.2)
    $: sidebarWidth = showSidebar ? `max(200px, ${$sidebarSize * 100}%)` : undefined
</script>

<section>
    <div class="sidebar" class:open={showSidebar} style:min-width={sidebarWidth} style:max-width={sidebarWidth}>
        {#if showSidebar && !$treeOrError.loading && $treeOrError.data}
            <FileTree
                activeEntry={$page.params.path ? last($page.params.path.split('/')) : ''}
                treeOrError={$treeOrError.data}
            >
                <h3 slot="title">
                    Files
                    <button on:click={() => (showSidebar = false)}
                        ><Icon svgPath={mdiChevronDoubleLeft} inline /></button
                    >
                </h3>
            </FileTree>
        {/if}
        {#if !showSidebar}
            <button class="open-sidebar" on:click={() => (showSidebar = true)}
                ><Icon svgPath={mdiChevronDoubleRight} inline /></button
            >
        {/if}
    </div>
    {#if showSidebar}
        <Separator currentPosition={sidebarSize} />
    {/if}
    <div class="content">
        <slot />
    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        overflow: hidden;
        margin: 1rem;
        margin-bottom: 0;
        flex: 1;
    }

    .sidebar {
        &.open {
            width: 200px;
        }

        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .content {
        flex: 1;
        margin: 0 1rem;
        background-color: var(--code-bg);
        overflow: hidden;
        display: flex;
        flex-direction: column;
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius);
    }

    button {
        border: 0;
        background-color: transparent;
        padding: 0;
        margin: 0;
        cursor: pointer;
    }

    h3 {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .open-sidebar {
        position: absolute;
        left: 0;
        border: 1px solid var(--border-color);
    }
</style>
