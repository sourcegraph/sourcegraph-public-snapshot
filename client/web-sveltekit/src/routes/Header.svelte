<script lang="ts">
    import { mdiBookOutline, mdiChartBar, mdiMagnify } from '@mdi/js'

    import { mark, svelteLogoEnabled } from '$lib/images'
    import type { AuthenticatedUser } from '$lib/shared'

    import HeaderNavLink from './HeaderNavLink.svelte'
    import { Button } from '$lib/wildcard'
    import UserMenu from './UserMenu.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { page } from '$app/stores'
    import CodyIcon from '$lib/icons/Cody.svelte'
    import CodeMonitoringIcon from '$lib/icons/CodeMonitoring.svelte'
    import BatchChangesIcon from '$lib/icons/BatchChanges.svelte'

    export let authenticatedUser: AuthenticatedUser | null | undefined

    $: reactURL = (function (url) {
        const urlCopy = new URL(url)
        urlCopy.searchParams.delete('feat')
        for (let feature of urlCopy.searchParams.getAll('feat')) {
            if (feature !== 'enable-sveltekit') {
                urlCopy.searchParams.append('feat', feature)
            }
        }
        urlCopy.searchParams.append('feat', '-enable-sveltekit')
        return urlCopy.toString()
    })($page.url)
</script>

<header>
    <a class="logo" href="/search">
        <img src={mark} alt="Sourcegraph" width="25" height="25" />
    </a>
    <nav class="ml-2">
        <ul>
            <HeaderNavLink href="/search" svgIconPath={mdiMagnify}>Code search</HeaderNavLink>
            <HeaderNavLink external href="/cody/chat">
                <CodyIcon slot="icon" />
                Cody
            </HeaderNavLink>
            <HeaderNavLink external href="/notebooks" svgIconPath={mdiBookOutline}>Notebooks</HeaderNavLink>
            <HeaderNavLink external href="/code-monitoring">
                <CodeMonitoringIcon slot="icon" />
                Monitoring
            </HeaderNavLink>
            <HeaderNavLink external href="/batch-changes">
                <BatchChangesIcon slot="icon" />
                Batch Changes
            </HeaderNavLink>
            <HeaderNavLink external href="/insights" svgIconPath={mdiChartBar}>Insights</HeaderNavLink>
        </ul>
    </nav>
    <Tooltip tooltip="Disable SvelteKit (go to React)">
        <a class="app-toggle" href={reactURL} data-sveltekit-reload>
            <img src={svelteLogoEnabled} alt="Svelte logo" width="20" height="20" />
        </a>
    </Tooltip>
    <div>
        {#if authenticatedUser}
            <UserMenu {authenticatedUser} />
        {:else}
            <Button variant="secondary" outline>
                <a slot="custom" let:className class={className} href="/sign-in" data-sveltekit-reload>Sign in</a>
            </Button>
        {/if}
    </div>
</header>

<style lang="scss">
    header {
        display: flex;
        align-items: center;
        border-bottom: 1px solid var(--border-color-2);
        height: var(--navbar-height);
        min-height: 40px;
        padding: 0 0.5rem;
        background-color: var(--color-bg-1);
    }

    .logo img {
        &:hover {
            @keyframes spin {
                50% {
                    transform: rotate(180deg) scale(1.2);
                }
                100% {
                    transform: rotate(180deg) scale(1);
                }
            }

            @media (prefers-reduced-motion: no-preference) {
                animation: spin 0.5s ease-in-out 1;
            }
        }
    }

    nav {
        display: flex;
        align-self: stretch;
        flex: 1;
    }

    ul {
        position: relative;
        padding: 0;
        margin: 0;
        display: flex;
        justify-content: center;
        list-style: none;
        background-size: contain;
    }

    .app-toggle {
        margin-right: 1rem;
    }
</style>
