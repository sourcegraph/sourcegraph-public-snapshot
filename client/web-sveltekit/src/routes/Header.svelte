<script lang="ts">
    import { mdiBookOutline, mdiChartBar, mdiMagnify } from '@mdi/js'

    import { mark } from '$lib/images'
    import type { AuthenticatedUser } from '$lib/shared'

    import HeaderNavLink from './HeaderNavLink.svelte'
    import { Button } from '$lib/wildcard'
    import UserMenu from './UserMenu.svelte'

    export let authenticatedUser: AuthenticatedUser | null | undefined
</script>

<header>
    <a href="/search">
        <img src={mark} alt="Sourcegraph" width="25" height="25" />
    </a>
    <nav class="ml-2">
        <ul>
            <HeaderNavLink href="/search" svgIconPath={mdiMagnify}>Code search</HeaderNavLink>
            <HeaderNavLink external href="/notebooks" svgIconPath={mdiBookOutline}>Notebooks</HeaderNavLink>
            <HeaderNavLink external href="/insights" svgIconPath={mdiChartBar}>Insights</HeaderNavLink>
        </ul>
    </nav>
    <div class="user">
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

    img {
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
</style>
