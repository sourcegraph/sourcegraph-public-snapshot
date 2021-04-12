import * as H from 'history'
import React from 'react'
import renderer, { act, ReactTestRenderer } from 'react-test-renderer'
import { of, NEVER } from 'rxjs'

import { FlatExtensionHostAPI } from '../api/contract'
import { ContributableMenu } from '../api/protocol'
import { pretendProxySubscribable, pretendRemote } from '../api/util'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { extensionsController } from '../util/searchTestHelpers'

import { ActionsNavItems } from './ActionsNavItems'

jest.mock('mdi-react/OpenInNewIcon', () => 'OpenInNewIcon')

describe('ActionItem', () => {
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined, settings: NEVER }
    const location = H.createLocation(
        'https://github.com/sourcegraph/sourcegraph/pull/5287/files#diff-eb9883bb910397a210512a13fd7384ac'
    )

    test('Renders contributed action items', async () => {
        let component!: ReactTestRenderer
        // eslint-disable-next-line @typescript-eslint/require-await
        await act(async () => {
            component = renderer.create(
                <ActionsNavItems
                    menu={ContributableMenu.EditorTitle}
                    location={location}
                    extensionsController={{
                        ...extensionsController,
                        extHostAPI: Promise.resolve(
                            pretendRemote<FlatExtensionHostAPI>({
                                getContributions: () =>
                                    pretendProxySubscribable(
                                        of({
                                            actions: [
                                                {
                                                    id: 'a',
                                                    actionItem: {
                                                        label: 'Action A',
                                                        description: 'This is Action A',
                                                    },
                                                },
                                            ],
                                            menus: {
                                                'editor/title': [
                                                    {
                                                        action: 'a',
                                                    },
                                                ],
                                            },
                                        })
                                    ),
                            })
                        ),
                    }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            )
        })
        expect(component.toJSON()).toMatchSnapshot()
    })
})
