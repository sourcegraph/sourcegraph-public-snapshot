import { checkOk } from '@sourcegraph/http-client'

import { isDefaultSourcegraphUrl } from '../util/context'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const { origin, hostname } = new URL(bundleURL)
    // Include credentials when fetching extensions from the private registry
    const includeCredentials = !isDefaultSourcegraphUrl(origin) && hostname !== 'localhost'
    const response = await fetch(bundleURL, {
        credentials: includeCredentials ? 'include' : 'omit',
    })
    checkOk(response)
    const blob = await response.blob()
    return window.URL.createObjectURL(blob)
}
