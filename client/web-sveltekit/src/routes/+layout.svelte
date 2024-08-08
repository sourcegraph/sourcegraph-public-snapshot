<script lang="ts">
    import { onDestroy } from 'svelte'
    import { writable } from 'svelte/store'

    import { browser } from '$app/environment'
    import { beforeNavigate } from '$app/navigation'
    import GlobalHeader from '$lib/navigation/GlobalHeader.svelte'
    import { TemporarySettingsStorage } from '$lib/shared'
    import { isLightTheme, setAppContext } from '$lib/stores'
    import { createTemporarySettingsStorage } from '$lib/temporarySettings'

    // When adding global imports here, they should probably also be added in .storybook/preview.ts
    import '@fontsource-variable/roboto-mono'
    import '@fontsource-variable/inter'
    import './styles.scss'

    import { isErrorLike } from '$lib/common'
    import { createFeatureFlagStore } from '$lib/featureflags'
    import FuzzyFinderContainer from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import GlobalNotification from '$lib/global-notifications/GlobalNotifications.svelte'
    import { getGraphQLClient } from '$lib/graphql/apollo'
    import { isRouteEnabled } from '$lib/navigation'

    import type { LayoutData } from './$types'
    import WelcomeOverlay from './WelcomeOverlay.svelte'

    export let data: LayoutData

    const user = writable(data.user ?? null)
    const settings = writable(isErrorLike(data.settings) ? null : data.settings)
    // It's OK to set the temporary storage during initialization time because
    // sign-in/out currently performs a full page refresh
    const temporarySettingsStorage = createTemporarySettingsStorage(
        data.user
            ? new TemporarySettingsStorage(getGraphQLClient(), true)
            : // Logged out storage
              new TemporarySettingsStorage(null, false)
    )

    setAppContext({
        user,
        settings,
        temporarySettingsStorage,
        featureFlags: createFeatureFlagStore(data.featureFlags, data.fetchEvaluatedFeatureFlags),
    })

    // We need to manually subscribe instead of using $isLightTheme because
    // at the moment Svelte tries to automatically subscribe to the store
    // the app context is not yet set.
    let lightTheme = false
    onDestroy(isLightTheme.subscribe(value => (lightTheme = value)))

    // Update stores when data changes
    $: $user = data.user ?? null
    $: $settings = isErrorLike(data.settings) ? null : data.settings

    $: if (browser) {
        document.documentElement.classList.toggle('theme-light', lightTheme)
        document.documentElement.classList.toggle('theme-dark', !lightTheme)
    }

    // Redirect the user to the react app when they navigate to a page that is
    // supported but not enabled.
    // (Routes that are not supported, i.e. don't exist in `routes/` are already
    // handled by SvelteKit (by triggering a browser refresh)).
    beforeNavigate(navigation => {
        if (navigation.willUnload || !navigation.to) {
            // Nothing to do here, request is already handled by the server
            return
        }

        if (isRouteEnabled(navigation.to.url.pathname)) {
            // Routes are handled by SvelteKit
            return
        }

        // Trigger page refresh to fetch the React app from the server
        navigation.cancel()
        window.location.href = navigation.to.url.toString()
    })

    $: currentUserID = data.user?.id
    $: handleOptOut = currentUserID
        ? async (): Promise<void> => {
              // Show departure message after switching off
              $temporarySettingsStorage.set('webNext.departureMessage.show', true)
              await data.disableSvelteFeatureFlags(currentUserID)
              window.location.reload()
          }
        : undefined
</script>

<svelte:head>
    <meta name="description" content="Code search" />
</svelte:head>

{#await data.globalSiteAlerts then globalSiteAlerts}
    {#if globalSiteAlerts}
        <GlobalNotification globalAlerts={globalSiteAlerts} />
    {/if}
{/await}

<GlobalHeader authenticatedUser={$user} {handleOptOut} entries={data.navigationEntries} />

<main>
    <slot />
</main>

<WelcomeOverlay />

<FuzzyFinderContainer />

<style lang="scss">
    :global(body) {
        height: 100vh;
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    main {
        isolation: isolate;
        flex: 1;
        display: flex;
        flex-direction: column;
        box-sizing: border-box;
        overflow-y: auto;
    }
</style>
