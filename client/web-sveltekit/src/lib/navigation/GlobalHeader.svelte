<script context="module" lang="ts">
    import { type Writable, writable, readonly } from 'svelte/store'

    export enum NavigationMode {
        PlainNavigation = 'global',
        WithCustomContent = 'with-custom-content',
    }

    const extensionElement: Writable<HTMLElement | null> = writable(null)

    export const navigationExtensionElement = readonly(extensionElement)
    export const navigationModeStore = writable<NavigationMode | `${NavigationMode}`>(NavigationMode.PlainNavigation)
</script>

<script lang="ts">
    import { mdiMenu } from '@mdi/js'

    import { browser } from '$app/environment'
    import Icon from '$lib/Icon.svelte'
    import { mark } from '$lib/images'
    import GlobalSidebarNavigation from '$lib/navigation/GlobalSidebarNavigation.svelte'
    import { mainNavigation } from '$lib/navigation/mainNavigation'
    import MainNavigationEntry from '$lib/navigation/MainNavigationEntry.svelte'
    import Popover from '$lib/Popover.svelte'
    import { Badge, Button } from '$lib/wildcard'

    import { GlobalNavigation_User } from './GlobalNavigation.gql'
    import UserMenu from './UserMenu.svelte'

    export let authenticatedUser: GlobalNavigation_User | null | undefined
    export let handleOptOut: (() => Promise<void>) | undefined

    let isSidebarNavigationOpen: boolean = false

    const isDevOrS2 =
        (browser && window.location.hostname === 'localhost') ||
        window.location.hostname === 'sourcegraph.sourcegraph.com'

    $: withCustomContent = $navigationModeStore === NavigationMode.WithCustomContent
</script>

<header class="root" data-global-header>
    {#if isSidebarNavigationOpen}
        <GlobalSidebarNavigation onClose={() => (isSidebarNavigationOpen = false)} />
    {/if}

    <div class="logo" class:with-custom-content={withCustomContent}>
        {#if withCustomContent}
            <button class="menu-button" on:click={() => (isSidebarNavigationOpen = true)}>
                <Icon svgPath={mdiMenu} aria-label="Navigation menu" />
            </button>
        {/if}

        <a href="/search">
            <img src={mark} alt="Sourcegraph" width="25" height="25" />
        </a>
    </div>

    <nav class="plain-navigation" bind:this={$extensionElement}>
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
        border-bottom: 1px solid var(--border-color);
        background-color: var(--color-bg-1);

        position: relative;
        height: 50px;
        min-width: 0;
    }

    .logo {
        display: flex;
        align-items: center;
        margin-left: 0.5rem;
        gap: 0.5rem;

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

        // Ensure that any content inside navigation portal block
        // can't overlap any static content like sidebar navigation
        // in the global header layout.
        isolation: isolate;

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
        display: flex;
        padding: 0.35rem;
        align-items: center;
        border: none;
        background-color: transparent;
        border-radius: var(--border-radius);

        &:hover {
            background-color: var(--secondary-2);
        }

        --icon-size: 1rem;
        --icon-fill-color: var(--icon-color);
    }
</style>
