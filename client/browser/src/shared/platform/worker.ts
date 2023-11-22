import { checkOk } from '@sourcegraph/http-client'

import { isDefaultSourcegraphUrl } from '../util/context'

/*
 * See [freeze_legacy_extensions.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@e8abca577d556b557372518170aaf045093ea760/-/blob/cmd/frontend/internal/registry/scripts/freeze_legacy_extensions.go)
 * and [sourcegraph/pull/45923](https://github.com/sourcegraph/sourcegraph/pull/45923) for more context.
 */
const LEGACY_EXTENSIONS_BUCKET_URL = 'https://storage.googleapis.com/sourcegraph-legacy-extensions/'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const { origin, hostname, href } = new URL(bundleURL)
    // Include credentials when fetching extensions from the private registry
    const includeCredentials =
        !isDefaultSourcegraphUrl(origin) && hostname !== 'localhost' && !href.startsWith(LEGACY_EXTENSIONS_BUCKET_URL)
    const response = await fetch(bundleURL, {
        credentials: includeCredentials ? 'include' : 'omit',
    })
    checkOk(response)
    const blob = await response.blob()
    return window.URL.createObjectURL(blob)
}
