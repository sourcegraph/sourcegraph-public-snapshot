<script lang="ts">
    import { mdiFlaskOutline } from '@mdi/js'

    import { mark } from '$lib/images'

    import { Button } from '$lib/wildcard'
    import UserMenu from './UserMenu.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { page } from '$app/stores'
    import type { Header_User } from './Header.gql'
    import Icon from '$lib/Icon.svelte'
    import { mainNavigation } from './mainNavigation'
    import MainNavigationEntry from './MainNavigationEntry.svelte'

    export let authenticatedUser: Header_User | null | undefined

    $: reactURL = (function (url) {
        const urlCopy = new URL(url)
        urlCopy.searchParams.delete('feat')
        for (let feature of urlCopy.searchParams.getAll('feat')) {
            if (feature !== 'web-next') {
                urlCopy.searchParams.append('feat', feature)
            }
        }
        urlCopy.searchParams.append('feat', '-web-next')
        return urlCopy.toString()
    })($page.url)
</script>

<header>
    <a class="logo" href="/search">
        <img src={mark} alt="Sourcegraph" width="25" height="25" />
    </a>
    <nav>
        <ul>
            {#each mainNavigation as entry (entry.label)}
                <MainNavigationEntry {entry} />
            {/each}
        </ul>
    </nav>
    <Tooltip tooltip="Leave experimental web app">
        <a href={reactURL} data-sveltekit-reload>
            <Icon svgPath={mdiFlaskOutline} --color="var(--oc-green-6)" />
        </a>
    </Tooltip>
    <div>
        {#if authenticatedUser}
            <UserMenu user={authenticatedUser} />
        {:else}
            <Button variant="secondary" outline>
                <svelte:fragment slot="custom" let:buttonClass>
                    <a class={buttonClass} href="/sign-in" data-sveltekit-reload>Sign in</a>
                </svelte:fragment>
            </Button>
        {/if}
    </div>
</header>

<style lang="scss">
    header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        height: var(--navbar-height);
        min-height: 40px;
        padding: 0 0.5rem;
        border-bottom: 1px solid var(--border-color-2);
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
        overflow-y: auto;
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
