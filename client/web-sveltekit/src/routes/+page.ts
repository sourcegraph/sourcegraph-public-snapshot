import { redirect } from '@sveltejs/kit'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = () => {
    redirect(300, '/search')
}
