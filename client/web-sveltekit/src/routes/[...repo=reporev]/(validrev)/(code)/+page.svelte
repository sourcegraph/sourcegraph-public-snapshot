<script lang="ts">
    import { mdiFileDocumentOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import { sidebarOpen } from '$lib/repo/stores'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const {
        value: readme,
        set: setReadme,
        pending: readmePending,
    } = createPromiseStore<PageData['deferred']['readme']>()
    $: setReadme(data.deferred.readme)
</script>

<h3 class="header">
    <div class="sidebar-button" class:hidden={$sidebarOpen}>
        <SidebarToggleButton />
    </div>
    {#if $readme}
        <Icon svgPath={mdiFileDocumentOutline} />
        &nbsp;
        {$readme.name}
    {:else if !$readmePending}
        Description
    {/if}
</h3>
<div class="content">
    {#if $readme?.richHTML}
        {@html $readme.richHTML}
    {:else if $readme?.content}
        <pre>{$readme.content}</pre>
    {:else if !$readmePending}
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

        :global(img) {
            max-width: 100%;
        }

        pre {
            white-space: pre-wrap;
        }
    }
</style>
