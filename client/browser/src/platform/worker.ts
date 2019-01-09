import { ajax } from 'rxjs/ajax'
import ExtensionHostWorker from 'worker-loader?inline!../../../../shared/src/api/extension/main.worker.ts'

export function createExtensionHostWorker(): ExtensionHostWorker {
    return new ExtensionHostWorker()
}

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const req = await ajax({
        url: bundleURL,
        crossDomain: true,
        responseType: 'blob',
    }).toPromise()
    return window.URL.createObjectURL(req.response)
}
