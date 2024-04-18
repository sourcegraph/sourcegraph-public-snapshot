<script context="module" lang="ts">
    import { type Writable, writable } from 'svelte/store'

    export enum NavigationMode {
        PlainNavigation = 'global',
        WithCustomContent = 'with-custom-content',
    }

    export const navigationModeStore = writable<NavigationMode | `${NavigationMode}`>(NavigationMode.PlainNavigation)
    export const navigationExtensionElement: Writable<HTMLElement | null> = writable(null)
</script>

<script lang="ts">
    import { mdiMenu } from '@mdi/js'
    import { browser } from '$app/environment'

    import { Badge, Button } from '$lib/wildcard'
    import { mark } from '$lib/images'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import { mainNavigation } from '$lib/navigation/mainNavigation'
    import GlobalSidebarNavigation from '$lib/navigation/GlobalSidebarNavigation.svelte'
    import MainNavigationEntry from '$lib/navigation/MainNavigationEntry.svelte'

    import UserMenu from './UserMenu.svelte'
    import { GlobalNavigation_User } from './GlobalNavigation.gql'

    export let authenticatedUser: GlobalNavigation_User | null | undefined
    export let handleOptOut: (() => Promise<void>) | undefined

    let isSidebarNavigationOpen: boolean = false

    function setSideNavigationState(open: boolean): void {
        isSidebarNavigationOpen = open
    }

    const isDevOrS2 =
        (browser && window.location.hostname === 'localhost') ||
        window.location.hostname === 'sourcegraph.sourcegraph.com'

    $: withCustomContent = $navigationModeStore === NavigationMode.WithCustomContent
</script>

<header class="root">
    {#if isSidebarNavigationOpen}
        <GlobalSidebarNavigation onClose={() => setSideNavigationState(false)} />
    {/if}

    {#if withCustomContent}
        <button class="menu-button" on:click={() => setSideNavigationState(true)}>
            <Icon svgPath={mdiMenu} aria-label="Navigation menu" />
        </button>
    {/if}

    <a class="logo" class:with-custom-content={withCustomContent} href="/search">
        <img src={mark} alt="Sourcegraph" width="25" height="25" />
    </a>

    <nav class="plain-navigation" bind:this={$navigationExtensionElement}>
        {#if !withCustomContent}
            <ul class="plain-navigation-list">
                {#each mainNavigation as entry (entry.label)}
                    <MainNavigationEntry {entry} />
                {/each}
            </ul>
        {/if}
    </nav>

    <Popover let:registerTrigger showOnHover>
        <span class="web-next-badge" use:registerTrigger>
            <Badge variant="warning">Experimental</Badge>
        </span>
        <div slot="content" class="web-next-content">
            <h3>Experimental web app</h3>
            <p>
                You are using an experimental version of the Sourcegraph web app. This version is under active
                development and may contain bugs or incomplete features.
            </p>
            {#if isDevOrS2}
                <p>
                    If you encounter any issues, please report them in our <a
                        href="https://sourcegraph.slack.com/archives/C05MHAP318B">Slack channel</a
                    >.
                </p>
            {/if}
            {#if handleOptOut}
                Or you can <button role="link" class="opt-out" on:click={handleOptOut}>opt out</button> of the Sveltekit
                experiment.
            {/if}
        </div>
    </Popover>
    <div>
        {#if authenticatedUser}
            <UserMenu user={authenticatedUser} />
        {:else}
            <Button variant="secondary" outline>
                <svelte:fragment slot="custom" let:buttonClass>
                    <a class={buttonClass} href="/sign-in">Sign in</a>
                </svelte:fragment>
            </Button>
        {/if}
    </div>
</header>

<style lang="scss">
    .root {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color-2);
        background-color: var(--color-bg-1);

        // This ensures that all arbitrary content is rendered above
        // other elements on the page.
        z-index: 1;
        position: relative;
        height: 50px;
        min-width: 0;
    }

    .logo {
        margin-left: 0.5rem;

        &.with-custom-content {
            margin-left: 0;
        }
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

    .plain-navigation {
        flex: 1;
        display: flex;
        align-self: stretch;
        min-width: 0;

        &-list {
            display: flex;
            gap: 1rem;
            padding: 0;
            margin: -0.5rem 0 -0.5rem 0;
            list-style: none;
            position: relative;
            justify-content: center;
            background-size: contain;
        }
    }

    // Opt out experiment badge and tooltip styles
    .opt-out {
        all: unset;
        cursor: pointer;
        color: var(--link-color);
        text-decoration: underline;
    }

    .web-next-badge {
        cursor: pointer;
        padding: 0.25rem;
    }

    .web-next-content {
        padding: 1rem;
        width: 20rem;

        p:last-child {
            margin-bottom: 0;
        }
    }

    // Custom menu with sidebar navigation controls styles
    .menu-button {
        border: none;
        padding: 0.35rem 0.35rem;
        border-radius: var(--border-radius);
        display: flex;
        align-items: center;
        background-color: transparent;
        margin-right: -0.5rem;

        &:hover {
            background-color: var(--secondary-2);
        }

        :global([data-icon]) {
            width: 1rem;
            height: 1rem;
            color: var(--icon-color);
        }
    }
</style>
