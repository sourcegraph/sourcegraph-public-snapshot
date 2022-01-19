import { BehaviorSubject } from 'rxjs'
import { filter, first, take } from 'rxjs/operators'
import sinon from 'sinon'
import sourcegraph from 'sourcegraph'

import { SettingsCascade } from '../../../settings/settings'
import { MainThreadAPI } from '../../contract'
import { Contributions } from '../../protocol'
import { pretendRemote } from '../../util'
import { activateExtensions, ExecutableExtension } from '../activation'
import { ExtensionHostState } from '../extensionHostState'

describe('Extension activation', () => {
    describe('activateExtensions()', () => {
        it('logs events for activated extensions', async () => {
            const logEvent = sinon.spy()

            const mockMain = pretendRemote<Pick<MainThreadAPI, 'getScriptURLForExtension' | 'logEvent'>>({
                getScriptURLForExtension: () => undefined,
                logEvent,
            })

            const FIXTURE_EXTENSION: ExecutableExtension = {
                scriptURL: 'https://fixture.extension',
                id: 'sourcegraph/fixture-extension',
                manifest: { url: 'a', contributes: {}, activationEvents: ['*'] },
            }

            const haveInitialExtensionsLoaded = new BehaviorSubject<boolean>(false)

            const mockState: Pick<
                ExtensionHostState,
                'activeExtensions' | 'contributions' | 'haveInitialExtensionsLoaded' | 'settings'
            > = {
                activeExtensions: new BehaviorSubject([FIXTURE_EXTENSION]),
                contributions: new BehaviorSubject<readonly Contributions[]>([]),
                haveInitialExtensionsLoaded,
                settings: new BehaviorSubject<Readonly<SettingsCascade>>({
                    subjects: [],
                    final: {},
                }),
            }

            // Noop for activation and deactivation
            const noopPromise = () => Promise.resolve()

            activateExtensions(
                mockState,
                mockMain,
                function createExtensionAPI() {
                    return {} as typeof sourcegraph
                },
                noopPromise,
                noopPromise
            )

            // Wait for extensions to load to check on the spy
            await haveInitialExtensionsLoaded
                .pipe(
                    filter(haveLoaded => haveLoaded),
                    first()
                )
                .toPromise()

            sinon.assert.calledWith(logEvent, 'ExtensionActivation', { extension_id: 'sourcegraph/fixture-extension' })
        })

        it('deactivates events when they no longer match the activation events', async () => {
            const mockMain = pretendRemote<Pick<MainThreadAPI, 'getScriptURLForExtension' | 'logEvent'>>({
                getScriptURLForExtension: () => undefined,
                logEvent: () => {},
            })

            const FIXTURE_EXTENSION: ExecutableExtension = {
                scriptURL: 'https://fixture.extension',
                id: 'sourcegraph/fixture-extension',
                manifest: { url: 'a', contributes: {}, activationEvents: ['*'] },
            }

            const haveInitialExtensionsLoaded = new BehaviorSubject<boolean>(false)
            const contributions = new BehaviorSubject<readonly Contributions[]>([])
            const activeExtensions = new BehaviorSubject([FIXTURE_EXTENSION])
            const mockStateWithActiveExtension: Pick<
                ExtensionHostState,
                'activeExtensions' | 'contributions' | 'haveInitialExtensionsLoaded' | 'settings'
            > = {
                activeExtensions,
                contributions,
                haveInitialExtensionsLoaded,
                settings: new BehaviorSubject<Readonly<SettingsCascade>>({
                    subjects: [],
                    final: {},
                }),
            }

            const onActivate = sinon.spy(() => Promise.resolve())
            const onDectivate = sinon.spy(() => Promise.resolve())

            activateExtensions(
                mockStateWithActiveExtension,
                mockMain,
                function createExtensionAPI() {
                    return {} as typeof sourcegraph
                },
                onActivate,
                onDectivate
            )

            // Wait for extensions to load to check on the spy
            await haveInitialExtensionsLoaded
                .pipe(
                    filter(haveLoaded => haveLoaded),
                    first()
                )
                .toPromise()

            sinon.assert.calledOnce(onActivate)
            sinon.assert.notCalled(onDectivate)

            activeExtensions.next([])

            // Wait for contributions to be propagated
            await contributions.pipe(take(2)).toPromise()

            sinon.assert.calledOnce(onActivate)
            sinon.assert.calledOnce(onDectivate)
        })
    })
})
