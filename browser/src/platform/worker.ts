import { ajax } from 'rxjs/ajax'
import { DEFAULT_SOURCEGRAPH_URL } from '../shared/util/context'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const req = await ajax({
        url: bundleURL,
        // Include credentials when fetching extensions from the private registry
        withCredentials: new URL(bundleURL).origin !== DEFAULT_SOURCEGRAPH_URL,
        crossDomain: true,
        responseType: 'blob',
    }).toPromise()
    return window.URL.createObjectURL(req.response)
}
