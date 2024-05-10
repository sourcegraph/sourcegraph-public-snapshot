<script context="module" lang="ts">
    import type { NavigationEntry, NavigationMenu } from './mainNavigation'

    function isNavigationMenu(entry: NavigationEntry | NavigationMenu): entry is NavigationMenu {
        return entry && 'children' in entry
    }
</script>

<script lang="ts">
    import { mdiClose } from '@mdi/js'

    import { page } from '$app/stores'
    import { onClickOutside, portal } from '$lib/dom'
    import Icon from '$lib/Icon.svelte'
    import SourcegraphLogo from '$lib/SourcegraphLogo.svelte'

    import { isCurrent } from './mainNavigation'
    import MainNavigationLink from './MainNavigationLink.svelte'

    export let onClose: () => void
    export let entries: (NavigationEntry | NavigationMenu)[]
</script>

<div class="root" use:portal>
    <div class="content" use:onClickOutside on:click-outside={onClose}>
        <header>
            <button class="close-button" on:click={onClose}>
                <Icon svgPath={mdiClose} aria-label="Close sidebar navigation" />
            </button>

            <a href="/search" class="logo-link">
                <SourcegraphLogo width="9.1rem" />
            </a>
        </header>

        <nav>
            <ul class="list">
                {#each entries as entry (entry.label)}
                    <li>
                        <MainNavigationLink {entry} />
                        {#if isNavigationMenu(entry) && entry.children.length > 0}
                            <ul>
                                {#each entry.children as subEntry (subEntry.label)}
                                    <li aria-current={isCurrent(subEntry, $page) ? 'page' : 'false'}>
                                        <MainNavigationLink entry={subEntry} />
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </li>
                {/each}
            </ul>
        </nav>
    </div>
</div>

<style lang="scss">
    .root {
        top: 0;
        left: 0;
        bottom: 0;
        right: 0;
        position: fixed;
        z-index: 1;
        // Fixed overlay color TODO: find a better design token for it
        background-color: rgba(172, 182, 192, 0.2);

        .content {
            display: flex;
            flex-direction: column;
            width: 18.75rem;
            height: 100%;
            transform: unset;
            border: none;
            padding: 0;
            background-color: var(--color-bg-1);
        }

        .close-button {
            border: none;
            padding: 0.35rem 0.35rem;
            border-radius: var(--border-radius);
            display: flex;
            align-items: center;
            background-color: transparent;

            &:hover {
                background-color: var(--secondary-2);
            }

            --icon-size: 1rem;
            --icon-fill-color: var(--icon-color);
        }

        header {
            display: flex;
            gap: 0.5rem;
            align-items: center;
            padding: 0.5rem;
            // Original menu navigation has 50px - 1px bottom border
            // To ensure that there are no jumps between closed/open states
            // we set height here to repeat menu and icon buttons positions.
            min-height: 49px;
        }

        .logo-link {
            flex-grow: 1;
            display: flex;
            align-items: center;
        }

        nav {
            padding-top: 1rem;
            padding-bottom: 1rem;
            flex-grow: 1;
            overflow: auto;
        }

        ul {
            padding: 0;
            margin: 0;
            list-style: none;
            display: flex;
            flex-direction: column;
            gap: 0.25rem;
            flex-grow: 1;

            li[aria-current='page'] {
                background-color: var(--secondary-2);
            }

            :global(a) {
                display: flex;
                flex-wrap: wrap;
                align-items: center;
                gap: 0.25rem;
                padding: 0.375rem 0.75rem;
                font-size: 1rem;

                --icon-fill-color: var(--icon-color);

                &:hover {
                    background-color: var(--secondary-2);
                }
            }

            & ul {
                :global(a) {
                    padding-left: 3.7rem;
                }
            }
        }
    }
</style>
