import { redirect } from '@sveltejs/kit'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = () => {
    // eslint-disable-next-line etc/throw-error, rxjs/throw-error
    throw redirect(300, '/search')
}
