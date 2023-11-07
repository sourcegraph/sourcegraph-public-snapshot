import { afterEach, beforeEach, describe, it } from '@jest/globals'
import { BehaviorSubject } from 'rxjs'
import sinon from 'sinon'

import { logger } from '@sourcegraph/common'

import type { ClientAPI } from '../../client/api/api'
import { pretendRemote } from '../../util'
import { proxySubscribable } from '../api/common'

import { initializeExtensionHostTest } from './test-helpers'

const noopMain = pretendRemote<ClientAPI>({
    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
    logExtensionMessage: (...data) => logger.log(...data),
})

describe('Extension logging', () => {
    let spy: sinon.SinonSpy
    beforeEach(() => {
        spy = sinon.spy(console, 'log')
    })

    afterEach(() => {
        spy.restore()
    })

    it('does not log when extension ID is absent from settings', () => {
        const extensionID = 'test/extension'

        const { extensionAPI } = initializeExtensionHostTest(
            {
                initialSettings: {
                    subjects: [],
                    final: {
                        'extensions.activeLoggers': [],
                    },
                },
                clientApplication: 'sourcegraph',
                sourcegraphURL: 'https://example.com/',
            },
            noopMain,
            extensionID
        )

        extensionAPI.app.log('message from extension')

        sinon.assert.notCalled(spy)
    })
    it('prefixes logs with extension ID', () => {
        const extensionID = 'test/extension'

        const { extensionAPI } = initializeExtensionHostTest(
            {
                initialSettings: {
                    subjects: [],
                    final: {
                        'extensions.activeLoggers': [extensionID],
                    },
                },
                clientApplication: 'sourcegraph',
                sourcegraphURL: 'https://example.com/',
            },
            noopMain,
            extensionID
        )

        extensionAPI.app.log('message from extension')

        sinon.assert.calledOnceWithExactly(
            spy,
            `🧩 %c${extensionID}`,
            'background-color: lightgrey;',
            'message from extension'
        )
    })
})
