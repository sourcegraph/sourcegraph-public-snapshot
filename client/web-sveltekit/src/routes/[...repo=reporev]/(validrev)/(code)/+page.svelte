<script lang="ts">
    // @sg RepoRoot
    import { onMount } from 'svelte'

    import { sidebarOpen } from '$lib/repo/stores'
    import { createPromiseStore } from '$lib/utils'
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'
    import Readme from '$lib/repo/Readme.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'

    import type { PageData } from './$types'
    import type { RepoPage_Readme } from './page.gql'

    export let data: PageData

    const readme = createPromiseStore<RepoPage_Readme | null>()
    $: readme.set(data.readme)

    onMount(() => {
        SVELTE_LOGGER.logViewEvent(SVELTE_TELEMETRY_EVENTS.ViewRepositoryPage)
    })
</script>

<svelte:head>
    <title>{data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<h3 class="header">
    <div class="sidebar-button" class:hidden={$sidebarOpen}>
        <SidebarToggleButton />
    </div>
    {#if $readme.value}
        {$readme.value.name}
    {:else if !$readme.pending}
        Description
    {/if}
</h3>
<div class="content">
    <div class="inner">
        {#if $readme.value}
            <Readme file={$readme.value} />
        {:else if !$readme.pending}
            {data.resolvedRevision.repo.description}
        {/if}
    </div>
</div>

<style lang="scss">
    h3 {
        margin: 0;
    }

    .header {
        position: sticky;
        top: 0;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        display: flex;
        align-items: center;
        background-color: var(--color-bg-1);

        .sidebar-button {
            margin-right: 0.5rem;

            // We still want the height of the button to be considered
            // when rendering the header, so that toggling the sidebar
            // won't change the height of the header.
            &.hidden {
                visibility: hidden;
                max-width: 0;
                margin-right: 0;
            }
        }
    }

    .content {
        overflow: auto;
        flex: 1;

        // We use an "inner" element to limit the width of the content while
        // keeping the scrollbar on the outer element, at the edge of the
        // viewport.
        .inner {
            max-width: var(--viewport-xl);
            padding: 1rem;
        }
    }
</style>
