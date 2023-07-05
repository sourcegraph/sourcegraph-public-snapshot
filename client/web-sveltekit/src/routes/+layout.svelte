<script lang="ts">
    import { setContext } from 'svelte'
    import { mdiBookOutline, mdiChartBar, mdiMagnify } from '@mdi/js'
    import { readable, writable, type Readable } from 'svelte/store'
    import { mark } from '$lib/images'

    import { browser } from '$app/environment'
    import { TemporarySettingsStorage } from '$lib/shared'
    import { KEY, type SourcegraphContext } from '$lib/stores'
    import { createTemporarySettingsStorage } from '$lib/temporarySettings'
    import HeaderNavLink from '$lib/HeaderNavLink.svelte'

    import './styles.scss'

    import { beforeNavigate } from '$app/navigation'

    import type { LayoutData, Snapshot } from './$types'

    export let data: LayoutData

    function createLightThemeStore(): Readable<boolean> {
        if (browser) {
            const matchMedia = window.matchMedia('(prefers-color-scheme: dark)')
            return readable(!matchMedia.matches, set => {
                const listener = (event: MediaQueryListEventMap['change']) => {
                    set(!event.matches)
                }
                matchMedia.addEventListener('change', listener)
                return () => matchMedia.removeEventListener('change', listener)
            })
        }
        return readable(true)
    }

    const user = writable(data.user ?? null)
    const settings = writable(null)
    const isLightTheme = createLightThemeStore()
    // It's OK to set the temporary storage during initialization time because
    // sign-in/out currently performs a full page refresh
    const temporarySettingsStorage = createTemporarySettingsStorage(
        data.user ? new TemporarySettingsStorage(data.graphqlClient, true) : undefined
    )

    setContext<SourcegraphContext>(KEY, {
        user,
        settings,
        isLightTheme,
        temporarySettingsStorage,
    })

    // Update stores when data changes
    $: $user = data.user ?? null

    $: if (browser) {
        document.documentElement.classList.toggle('theme-light', $isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !$isLightTheme)
    }

    let main: HTMLElement | null = null
    let scrollTop = 0
    beforeNavigate(() => {
        // It looks like `snapshot.capture` is called "too late", i.e. after the
        // content has been updated. beforeNavigate is used to capture the correct
        // scroll offset
        scrollTop = main?.scrollTop ?? 0
    })
    export const snapshot: Snapshot<{ x: number }> = {
        capture() {
            return { x: scrollTop }
        },
        restore(value) {
            const start = Date.now()
            requestAnimationFrame(function scroll() {
                if (main) {
                    main.scrollTo(0, value.x)
                }
                if ((!main || main.scrollTop !== value.x) && Date.now() - start < 3000) {
                    requestAnimationFrame(scroll)
                }
            })
        },
    }
</script>

<svelte:head>
    <title>Sourcegraph</title>
    <meta name="description" content="Code search" />
</svelte:head>

<div class="app">
    <div class="dock">
        <div class="bar">
        <img src={mark} class="mt-2" alt="Sourcegraph" width="25" height="25" />
        <nav class="ml-2">
            <ul>
                <HeaderNavLink href="/search" svgIconPath={mdiMagnify}><span>Code search</span></HeaderNavLink>
                <HeaderNavLink href="/notebooks" svgIconPath={mdiBookOutline}><span>Notebooks</span></HeaderNavLink>
                <HeaderNavLink href="/insights" svgIconPath={mdiChartBar}><span>Insights</span></HeaderNavLink>
            </ul>
        </nav>
    </div>
    </div>
    <main bind:this={main}>
        <slot />
    </main>
</div>

<style lang="scss">
    .app {
        display: flex;
        height: 100vh;
        overflow: hidden;

    }

    .dock {
        width: 40px;
        background-color: black;
        display: flex;
        flex-direction:  column;
        align-items: left;


        img {
            align-self: center;
        }

        ul {
            padding: 0;
            margin: 0;
            list-style: none;
            background-size: contain;

            span {
                display: none;
            }

        }

        .bar {
            display: flex;
        flex-direction:  column;
            flex: 1;
            width: 40px;
            transition: width 100ms ease-in-out;
        overflow: hidden;

            &:hover {
            transition: width 100ms ease-in-out;
            box-shadow: 5px 0px 10px -7px rgba(0,0,0,0.75);
                background-color: black;
                    z-index: 100;
                    position: fixed;
                    top: 0;
                    bottom: 0;
                    width: 150px;
                    span {
                        display:initial;
                    }
            }
        }

    }

    main {
        flex: 1;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
        overflow: auto;
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


</style>
