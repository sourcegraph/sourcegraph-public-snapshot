<script lang="ts">
    // @sg RepoRoot EnableRollout
    import { onMount } from 'svelte'

    import Readme from '$lib/repo/Readme.svelte'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'
    import type { RepoPage_Readme } from './page.gql'
    import OpenCodyAction from '$lib/repo/OpenCodyAction.svelte'
    import MobileFileSidePanelOpenButton from '$lib/repo/MobileFileSidePanelOpenButton.svelte'

    export let data: PageData

    const readme = createPromiseStore<RepoPage_Readme | null>()
    $: readme.set(data.readme)
    $: isCodyAvailable = data.isCodyAvailable

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('repo', 'view')
    })
</script>

<svelte:head>
    <title>{data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<div class="header">
    <MobileFileSidePanelOpenButton />

    <h3>
        {#if $readme.value}
            {$readme.value.name}
        {:else if !$readme.pending}
            Description
        {/if}
    </h3>
    <div class="actions">
        {#if $isCodyAvailable}
            <OpenCodyAction />
        {/if}
    </div>
</div>
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
    .header {
        position: sticky;
        top: 0;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        background-color: var(--color-bg-1);
        height: var(--repo-header-height);

        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    h3 {
        margin: 0;
        flex: 1;
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
