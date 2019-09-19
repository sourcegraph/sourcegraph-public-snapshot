import { DEFAULT_SOURCEGRAPH_URL } from '../shared/util/context'
import { checkOk } from '../../../shared/src/backend/fetch'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const response = await fetch(bundleURL, {
        // Include credentials when fetching extensions from the private registry
        credentials: new URL(bundleURL).origin !== DEFAULT_SOURCEGRAPH_URL ? 'include' : 'omit',
    })
    checkOk(response)
    const blob = await response.blob()
    return window.URL.createObjectURL(blob)
}
