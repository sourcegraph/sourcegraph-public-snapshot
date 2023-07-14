<script lang="ts">
    import { setContext } from 'svelte'
    import { readable, writable, type Readable } from 'svelte/store'

    import { browser } from '$app/environment'
    import { isErrorLike } from '$lib/common'
    import { TemporarySettingsStorage } from '$lib/shared'
    import { KEY, scrollAll, type SourcegraphContext } from '$lib/stores'
    import { createTemporarySettingsStorage } from '$lib/temporarySettings'

    import Header from './Header.svelte'

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
    const settings = writable(isErrorLike(data.settings) ? null : data.settings.final)
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
    $: $settings = isErrorLike(data.settings) ? null : data.settings.final

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
            restoreScrollPosition(value.x)
        },
    }

    function restoreScrollPosition(y: number) {
        const start = Date.now()
        requestAnimationFrame(function scroll() {
            if (main) {
                main.scrollTo(0, y)
            }
            if ((!main || main.scrollTop !== y) && Date.now() - start < 3000) {
                requestAnimationFrame(scroll)
            }
        })
    }
</script>

<svelte:head>
    <title>Sourcegraph</title>
    <meta name="description" content="Code search" />
</svelte:head>

<div class="app" class:overflowHidden={!$scrollAll}>
    <Header authenticatedUser={$user} />

    <main bind:this={main}>
        <slot />
    </main>
</div>

<style lang="scss">
    .app {
        display: flex;
        flex-direction: column;
        height: 100vh;
        overflow-y: auto;

        &.overflowHidden {
            overflow: hidden;

            main {
                overflow-y: auto;
            }
        }
    }

    main {
        flex: 1;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
    }
</style>
