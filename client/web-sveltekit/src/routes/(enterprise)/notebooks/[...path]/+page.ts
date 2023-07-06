import { createPlatformContext } from '$lib/context'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => {
    return {
        platformContext: parent().then(({ graphqlClient }) => createPlatformContext(graphqlClient)),
    }
}
