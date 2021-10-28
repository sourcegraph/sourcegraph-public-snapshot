import { checkOk } from '@sourcegraph/shared/src/backend/fetch'

import { DEFAULT_SOURCEGRAPH_URL } from '../util/context'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const { origin, hostname } = new URL(bundleURL)
    // Include credentials when fetching extensions from the private registry
    const includeCredentials = origin !== DEFAULT_SOURCEGRAPH_URL && hostname !== 'localhost'
    const response = await fetch(bundleURL, {
        credentials: includeCredentials ? 'include' : 'omit',
    })
    checkOk(response)
    const blob = await response.blob()
    return window.URL.createObjectURL(blob)
}
