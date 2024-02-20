<script lang="ts">
    import { setContext } from 'svelte'
    import { writable } from 'svelte/store'

    import { browser } from '$app/environment'
    import { isErrorLike } from '$lib/common'
    import { TemporarySettingsStorage } from '$lib/shared'
    import { isLightTheme, KEY, scrollAll, type SourcegraphContext } from '$lib/stores'
    import { createTemporarySettingsStorage, temporarySetting } from '$lib/temporarySettings'
    import { setThemeFromString } from '$lib/theme'
    import { classNames } from '$lib/dom'

    import Header from './Header.svelte'

    import './styles.scss'

    import type { LayoutData } from './$types'
    import { createFeatureFlagStore } from '$lib/featureflags'
    import InfoBanner from './InfoBanner.svelte'
    import { getGraphQLClient } from '$lib/graphql/apollo'

    export let data: LayoutData

    const user = writable(data.user ?? null)
    const settings = writable(isErrorLike(data.settings) ? null : data.settings)
    // It's OK to set the temporary storage during initialization time because
    // sign-in/out currently performs a full page refresh
    const temporarySettingsStorage = createTemporarySettingsStorage(
        data.user ? new TemporarySettingsStorage(getGraphQLClient(), true) : undefined
    )

    setContext<SourcegraphContext>(KEY, {
        user,
        settings,
        temporarySettingsStorage,
        featureFlags: createFeatureFlagStore(data.featureFlags, data.fetchEvaluatedFeatureFlags),
    })

    // Update stores when data changes
    $: $user = data.user ?? null
    $: $settings = isErrorLike(data.settings) ? null : data.settings

    // Set initial, user configured theme
    // TODO: This should be send be server in the HTML so that we don't flash the wrong theme
    // on initial page load.
    $: userTheme = temporarySetting('user.themePreference', 'System')
    $: if (!$userTheme.loading && $userTheme.data) {
        setThemeFromString($userTheme.data)
    }

    $: if (browser) {
        document.documentElement.classList.toggle('theme-light', $isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !$isLightTheme)
    }
</script>

<svelte:head>
    <title>Sourcegraph</title>
    <meta name="description" content="Code search" />
</svelte:head>

<svelte:body use:classNames={$scrollAll ? '' : 'overflowHidden'} />

<InfoBanner />
<Header authenticatedUser={$user} />

<main>
    <slot />
</main>

<style lang="scss">
    :global(body.overflowHidden) {
        display: flex;
        flex-direction: column;
        height: 100vh;
        overflow: hidden;

        main {
            overflow-y: auto;
        }
    }

    main {
        flex: 1;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
    }
</style>
