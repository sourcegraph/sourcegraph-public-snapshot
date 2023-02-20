<script lang="ts">
    import { onMount, setContext } from 'svelte'
    import { readable, writable } from 'svelte/store'

    import { browser } from '$app/environment'
    import { KEY, type SourcegraphContext } from '$lib/stores'
    import { isErrorLike } from '$lib/common'
    import { observeSystemIsLightTheme, TemporarySettingsStorage } from '$lib/shared'
    import { readableObservable } from '$lib/utils'
    import { createTemporarySettingsStorage } from '$lib/temporarySettings'

    import Header from './Header.svelte'
    import './styles.scss'
    import type { LayoutData } from './$types'
    import { setExperimentalFeaturesFromSettings } from '$lib/web'

    export let data: LayoutData

    const user = writable(data.user ?? null)
    const settings = writable(isErrorLike(data.settings) ? null : data.settings.final)
    const platformContext = writable(data.platformContext)
    const isLightTheme = browser ? readableObservable(observeSystemIsLightTheme(window).observable) : readable(true)
    // It's OK to set the temporary storage during initialization time because
    // sign-in/out currently performs a full page refresh
    const temporarySettingsStorage = createTemporarySettingsStorage(
        data.user ? new TemporarySettingsStorage(data.graphqlClient, true) : undefined
    )

    setContext<SourcegraphContext>(KEY, {
        user,
        settings,
        platformContext,
        isLightTheme,
        temporarySettingsStorage,
    })

    onMount(() => {
        // Settings can change over time. This ensures that the store is always
        // up-to-date.
        const settingsSubscription = data.platformContext?.settings.subscribe(newSettings => {
            settings.set(isErrorLike(newSettings.final) ? null : newSettings.final)
        })
        return () => settingsSubscription?.unsubscribe()
    })

    $: $user = data.user ?? null
    $: $settings = isErrorLike(data.settings) ? null : data.settings.final
    $: $platformContext = data.platformContext
    // Sync React stores
    $: setExperimentalFeaturesFromSettings(data.settings)


    $: if (browser) {
        document.documentElement.classList.toggle('theme-light', $isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !$isLightTheme)
    }
</script>

<svelte:head>
    <title>Sourcegraph</title>
    <meta name="description" content="Code search" />
</svelte:head>

<div class="app">
    <Header authenticatedUser={$user} />

    <main>
        <slot />
    </main>
</div>

<style lang="scss">
    .app {
        display: flex;
        flex-direction: column;
        height: 100vh;
        overflow: hidden;
    }

    main {
        flex: 1;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
        overflow: auto;
    }
</style>
