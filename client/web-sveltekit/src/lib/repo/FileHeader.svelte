<script lang="ts">
    import { page } from '$app/stores'

    import SidebarToggleButton from './SidebarToggleButton.svelte'
    import { sidebarOpen } from './stores'
    import { navFromPath } from './utils'

    $: breadcrumbs = navFromPath($page.params.path, $page.params.repo)
</script>

<div class="header">
    <div class="toggle-wrapper" class:hidden={$sidebarOpen}>
        <SidebarToggleButton />
    </div>
    <h2>
        {#each breadcrumbs as [name, path], index}
            <!--
                This space is necessary to enable wrapping of each breadcrumb
                if there is not enough space.
            -->
            {' '}
            <span class:last={index === breadcrumbs.length - 1}>
                {#if index > 0}
                    /
                {/if}
                {#if path}
                    <a href={path}>{name}</a>
                {:else}
                    <slot name="icon" />
                    {name}
                {/if}
            </span>
        {/each}
    </h2>
    <slot name="actions" />
</div>

<style lang="scss">
    .header {
        display: flex;
        align-items: baseline;
        padding: 0.25rem 0.5rem;
        border-bottom: 1px solid var(--border-color);
        position: sticky;
        top: 0px;
        background-color: var(--color-bg-1);
        z-index: 1;
        container: fileheader / inline-size;
        gap: 0.5rem;
    }

    .toggle-wrapper {
        margin-right: 0.5rem;

        // We still want the height of the button to be considered
        // when rendering the header, so that toggling the sidebar
        // won't change the height of the header.
        &.hidden {
            visibility: hidden;
            margin-right: 0;
            width: 0;
        }
    }

    h2 {
        font-weight: 400;
        font-size: var(--code-font-size);
        font-family: var(--code-font-family);
        margin: 0;

        a {
            color: var(--body-color);
        }

        span {
            white-space: nowrap;
        }

        .last {
            font-weight: bold;
        }
    }

    a {
        flex: 1;
    }
</style>
