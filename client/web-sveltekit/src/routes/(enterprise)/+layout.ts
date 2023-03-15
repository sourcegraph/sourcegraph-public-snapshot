import { error } from '@sveltejs/kit'

import { PUBLIC_SG_ENTERPRISE } from '$env/static/public'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = () => {
    // Example for how we could prevent access to all enterprese specific routes.
    // It's not quite the same as not having the routes at all and have the
    // interpreted differently, like in the current web app.
    if (!PUBLIC_SG_ENTERPRISE) {
        // eslint-disable-next-line etc/throw-error, rxjs/throw-error
        throw error(404, { message: 'enterprise feature' })
    }
}
