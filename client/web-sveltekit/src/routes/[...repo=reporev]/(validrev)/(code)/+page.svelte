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
    {#if $readme.value}
        <Readme file={$readme.value} />
    {:else if !$readme.pending}
        {data.resolvedRevision.repo.description}
    {/if}
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
        padding: 1rem;
        overflow: auto;
        flex: 1;
    }
</style>
