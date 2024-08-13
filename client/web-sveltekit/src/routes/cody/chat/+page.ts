import { redirect } from '@sveltejs/kit'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const dashboardRoute = window.context.sourcegraphDotComMode ? '/cody/manage' : '/cody/dashboard'
    const data = await parent()

    if (!data.user) {
        redirect(302, '/sign-in')
    }

    if (!window.context?.codyEnabledForCurrentUser) {
        redirect(303, dashboardRoute)
    }

    return {
        dashboardRoute,
    }
}
