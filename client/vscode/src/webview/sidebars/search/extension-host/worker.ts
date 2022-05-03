import { Subscription } from 'rxjs'

import { ClosableEndpointPair } from '@sourcegraph/shared/src/platform/context'

import { createEndpointsForWebToWeb } from '../../../comlink/webviewEndpoint'

// eslint-disable-next-line import/extensions
import ExtensionHostWorker from './main.worker.ts'

export function createExtensionHost(): ClosableEndpointPair {
    const worker = new ExtensionHostWorker()
    const { webview: expose, worker: proxy } = createEndpointsForWebToWeb(worker)

    return { endpoints: { expose, proxy }, subscription: new Subscription(() => worker.terminate()) }
}
