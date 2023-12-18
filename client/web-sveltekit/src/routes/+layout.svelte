<script lang="ts">
    import { setContext } from 'svelte'
    import { readable, writable } from 'svelte/store'

    import { browser } from '$app/environment'
    import { isErrorLike } from '$lib/common'
    import { TemporarySettingsStorage } from '$lib/shared'
    import { isLightTheme, KEY, scrollAll, type SourcegraphContext } from '$lib/stores'
    import { createTemporarySettingsStorage, temporarySetting } from '$lib/temporarySettings'
    import { humanTheme } from '$lib/theme'

    import Header from './Header.svelte'

    import './styles.scss'

    import { beforeNavigate } from '$app/navigation'

    import type { LayoutData, Snapshot } from './$types'
    import { createFeatureFlagStore, fetchEvaluatedFeatureFlags } from '$lib/featureflags'
    import InfoBanner from './InfoBanner.svelte'
    import { Theme } from '$lib/theme'

    export let data: LayoutData

    const user = writable(data.user ?? null)
    const settings = writable(isErrorLike(data.settings) ? null : data.settings.final)
    // It's OK to set the temporary storage during initialization time because
    // sign-in/out currently performs a full page refresh
    const temporarySettingsStorage = createTemporarySettingsStorage(
        data.user ? new TemporarySettingsStorage(data.graphqlClient, true) : undefined
    )

    setContext<SourcegraphContext>(KEY, {
        user,
        settings,
        temporarySettingsStorage,
        featureFlags: createFeatureFlagStore(data.featureFlags, fetchEvaluatedFeatureFlags),
        client: readable(data.graphqlClient),
    })

    // Update stores when data changes
    $: $user = data.user ?? null
    $: $settings = isErrorLike(data.settings) ? null : data.settings.final

    // Set initial, user configured theme
    // TODO: This should be send be server in the HTML so that we don't flash the wrong theme
    // on initial page load.
    $: userTheme = temporarySetting('user.themePreference', 'System')
    $: if (!$userTheme.loading && $userTheme.data) {
        $humanTheme = $userTheme.data
    }

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
    <InfoBanner />
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
