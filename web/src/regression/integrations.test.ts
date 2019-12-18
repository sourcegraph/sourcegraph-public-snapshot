/**
 * @jest-environment node
 */

import { getConfig } from '../../../shared/src/e2e/config'
import { fromFetch } from 'rxjs/fetch'
import { map, catchError } from 'rxjs/operators'
import { checkOk } from '../../../shared/src/backend/fetch'
import { merge } from 'rxjs'

describe('Native integrations regression test suite', () => {
    const { sourcegraphBaseUrl } = getConfig('sourcegraphBaseUrl')
    test('Native integration assets are served by the instance', async () => {
        const assets = [
            '/.assets/extension/scripts/integration.bundle.js',
            '/.assets/extension/scripts/phabricator.bundle.js',
            '/.assets/extension/scripts/extensionHostWorker.bundle.js',
            '/.assets/extension/css/style.bundle.css',
            '/.assets/extension/extensionHostFrame.html',
        ]
        await merge(
            ...assets.map(asset =>
                fromFetch(new URL(asset, sourcegraphBaseUrl).href).pipe(
                    map(checkOk),
                    catchError(() => {
                        throw new Error('Error fetching native integration asset: ' + asset)
                    })
                )
            )
        ).toPromise()
    })
})
