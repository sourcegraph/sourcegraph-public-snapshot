<script lang="ts">
    import { page } from '$app/stores'

    import Permalink from '../../routes/[...repo]/Permalink.svelte'

    import SidebarToggleButton from './SidebarToggleButton.svelte'
    import { sidebarOpen } from './stores'
    import { navFromPath } from './utils'

    $: breadcrumbs = navFromPath($page.params.path, $page.params.repo)
</script>

<div class="header">
    <div class="toggle-wrapper" class:open={$sidebarOpen}>
        <SidebarToggleButton />&nbsp;
    </div>
    <h2>
        <span class="icon">
            <slot name="icon" />&nbsp;
        </span>
        <span>
            {#each breadcrumbs as [name, path], index}
                {#if index > 0}
                    /
                {/if}
                <span class:last={index === breadcrumbs.length - 1}>
                    {#if path}
                        <a href={path}>{name}</a>
                    {:else}
                        {name}
                    {/if}
                </span>
            {/each}
        </span>
    </h2>
    <div class="actions">
        <slot name="actions" />
        <Permalink />
    </div>
</div>

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        position: sticky;
        top: 0px;
        background-color: var(--color-bg-1);
        z-index: 1;
    }

    .toggle-wrapper {
        display: initial;
        margin-right: 0.1rem;

        &.open {
            display: none;
        }
    }

    h2 {
        display: flex;
        align-items: center;
        font-weight: normal;
        font-size: 1rem;
        margin: 0;

        .last {
            font-weight: bold;
        }

        .icon {
            flex-shrink: 0;
        }
    }

    .actions {
        margin-left: auto;
    }

    a {
        flex: 1;
    }
</style>
