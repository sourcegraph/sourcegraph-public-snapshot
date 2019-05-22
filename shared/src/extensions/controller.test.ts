import { take } from 'lodash'
import { from, Observable } from 'rxjs'
import * as sinon from 'sinon'
import { ContributionRegistry } from '../api/client/services/contribution'
import { ExecutableExtension } from '../api/client/services/extensionsService'
import { Contributions } from '../api/protocol/contribution'
import { registerExtensionContributions } from './controller'
import { ExtensionManifest } from './extensionManifest'

describe('registerExtensionContributions()', () => {
    test('registers a contributions observable that emits contributions from all extensions', async () => {
        const registerContributions = sinon.spy()
        const fakeRegistry: Pick<ContributionRegistry, 'registerContributions'> = {
            registerContributions,
        }
        const extensions: ExecutableExtension[] = [
            {
                id: '1',
                scriptURL: '1.js',
                manifest: {
                    contributes: {
                        configuration: {
                            '1.config': '1',
                        },
                    },
                } as any,
            },
            {
                id: '2',
                scriptURL: '2.js',
                manifest: {
                    contributes: {
                        configuration: {
                            '2.config': '2',
                        },
                    },
                } as any,
            },
            {
                id: '3',
                scriptURL: '3.js',
                manifest: {
                    contributes: {
                        configuration: {
                            '3.config': '3',
                        },
                    },
                } as any,
            },
        ]
        registerExtensionContributions(fakeRegistry, {
            activeExtensions: from([take(extensions, 1), take(extensions, 2), extensions]),
        })
        sinon.assert.calledOnce(registerContributions)
        const extensionContributions: Observable<Contributions[]> = registerContributions.getCalls()[0].args[0]
            .contributions
        expect(await extensionContributions.toPromise()).toEqual(
            extensions.map(({ manifest }) => (manifest as NonNullable<ExtensionManifest>).contributes)
        )
    })
})
