import { error, redirect } from '@sveltejs/kit'

import { isErrorLike, parseJSONCOrError } from '$lib/common'
import { getGraphQLClient } from '$lib/graphql'
import type { SettingsEdit } from '$lib/graphql-types'
import type { Settings } from '$lib/shared'

import type { LayoutLoad } from './$types'
import {
    Init,
    EvaluatedFeatureFlagsQuery,
    GlobalAlertsSiteFlags,
    DisableSveltePrototype,
    EditSettings,
    LatestSettingsQuery,
} from './layout.gql'
import { getMainNavigationEntries, Mode } from './navigation'

// Disable server side rendering for the whole app
export const ssr = false
export const prerender = false

export const load: LayoutLoad = async ({ fetch }) => {
    const client = getGraphQLClient()

    // We don't block the whole page loader with site alerts
    // it's handled later in the page svelte template, render page
    // immediately as soon as we have init data
    const globalSiteAlerts = client.query(GlobalAlertsSiteFlags, {}, { fetch, requestPolicy: 'network-only' })

    const result = await client.query(Init, {}, { fetch, requestPolicy: 'network-only' })

    if (result.error?.response?.status === 401) {
        // The server will take care of redirecting to the sign-in page, but when
        // developing locally an proxying the API requests, we need to do it ourselves.
        redirect(307, '/sign-in')
    }

    if (!result.data || result.error) {
        error(500, `Failed to initialize app: ${result.error}`)
    }

    const settings = parseJSONCOrError<Settings>(result.data.viewerSettings.final)

    if (isErrorLike(settings)) {
        error(500, `Failed to parse user settings: ${settings.message}`)
    }

    return {
        navigationEntries: getMainNavigationEntries(
            (window.context.sourcegraphDotComMode ? Mode.DOTCOM : Mode.ENTERPRISE) |
                (window.context.codyEnabledOnInstance ? Mode.CODY_INSTANCE_ENABLED : 0) |
                (window.context.codyEnabledForCurrentUser ? Mode.CODY_USER_ENABLED : 0) |
                (window.context.batchChangesEnabled ? Mode.BATCH_CHANGES_ENABLED : 0) |
                (window.context.codeInsightsEnabled ? Mode.CODE_INSIGHTS_ENABLED : 0) |
                (result.data.currentUser ? Mode.AUTHENTICATED : Mode.UNAUTHENTICATED)
        ),

        // User data
        user: result.data.currentUser,
        settings,
        featureFlags: result.data.evaluatedFeatureFlags,

        globalSiteAlerts: globalSiteAlerts.then(result => result.data?.site),
        fetchEvaluatedFeatureFlags: async () => {
            const result = await client.query(EvaluatedFeatureFlagsQuery, {}, { requestPolicy: 'network-only', fetch })
            if (!result.data || result.error) {
                throw new Error(`Failed to fetch evaluated feature flags: ${result.error}`)
            }
            return result.data.evaluatedFeatureFlags
        },
        disableSvelteFeatureFlags: async (userID: string) => {
            const mutationResult = await client.mutation(
                DisableSveltePrototype,
                { userID },
                { requestPolicy: 'network-only', fetch }
            )
            if (!mutationResult.data || mutationResult.error) {
                throw new Error(`Failed to disable svelte feature flags: ${result.error}`)
            }
        },
        updateUserSetting: async (edit: SettingsEdit): Promise<void> => {
            // We have to set network-only here, because otherwise the client will reuse a previously cached value
            const latestSettings = await client.query(LatestSettingsQuery, {}, { requestPolicy: 'network-only', fetch })
            if (!latestSettings.data || latestSettings.error) {
                throw new Error(`Failed to fetch latest settings during editor update: ${latestSettings.error}`)
            }
            const userSetting = latestSettings.data.viewerSettings.subjects.find(s => s.__typename === 'User')
            if (!userSetting) {
                throw new Error('Failed to find user settings subject')
            }
            const lastID = userSetting.latestSettings?.id
            if (!lastID) {
                throw new Error('Failed to get new last ID from settings result')
            }
            const mutationResult = await client.mutation(
                EditSettings,
                {
                    lastID,
                    subject: userSetting.id,
                    edit,
                },
                { requestPolicy: 'network-only', fetch }
            )
            if (!mutationResult.data || mutationResult.error) {
                throw new Error(`Failed to update editor path: ${mutationResult.error}`)
            }
        },
    }
}
