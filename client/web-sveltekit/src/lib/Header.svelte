<script lang="ts">
    import { mdiBookOutline, mdiChartBar, mdiClose, mdiMagnify } from '@mdi/js'

    import UserAvatar from '$lib/UserAvatar.svelte'

    import HeaderNavLink from './HeaderNavLink.svelte'
    import Button from '$lib/wildcard/Button.svelte'
    import Icon from '$lib/Icon.svelte'
    import { user } from './stores'

    $: authenticatedUser = user

    let showNavigation = false
</script>

<header>
    <div class="header">
        <slot />
    </div>
    <div class="user">
        {#if $authenticatedUser}
            <UserAvatar user={$authenticatedUser} />
            <!--
                Needs data-sveltekit-reload to force hitting the server and
                proxying the request to the Sourcegraph instance.
            -->
            <a href="/-/sign-out" data-sveltekit-reload>Sign out</a>
        {:else}
            <a href="/sign-in" data-sveltekit-reload>Sign in</a>
        {/if}
    </div>
</header>


{#if showNavigation}
   <div class="navigation">
   <div class="content">
    <Button variant="icon" on:click={() => showNavigation = false}><Icon inline svgPath={mdiClose} /></Button>
    <nav class="ml-2">
        <ul>
            <HeaderNavLink href="/search" svgIconPath={mdiMagnify}>Code search</HeaderNavLink>
            <HeaderNavLink href="/notebooks" svgIconPath={mdiBookOutline}>Notebooks</HeaderNavLink>
            <HeaderNavLink href="/insights" svgIconPath={mdiChartBar}>Insights</HeaderNavLink>
        </ul>
    </nav>
    </div>
</div>
{/if}

<style lang="scss">
    .navigation {
        z-index: 100;
        position: fixed;
        left: 0;
        right: 0;
        top: 0;
        bottom: 0;
        background-color: var(--modal-bg);


       .content {
            width: 300px;
            height: 100vh;
            background-color: var(--color-bg-1);
        }
    }


    .header {
        display: flex;
        align-self: center;
        flex: 1;
    }

    header {
        display: flex;
        align-items: center;
        border-bottom: 1px solid var(--border-color-2);
        height: var(--navbar-height);
        min-height: 40px;
        padding: 0 1rem;
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

    ul {
        position: relative;
        padding: 0;
        margin: 0;
        list-style: none;
        background-size: contain;
    }
</style>
