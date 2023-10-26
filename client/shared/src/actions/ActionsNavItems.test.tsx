import { describe, expect, test, jest } from '@jest/globals'
import { render, act, type RenderResult } from '@testing-library/react'
import * as H from 'history'
import { of, NEVER } from 'rxjs'

import { ContributableMenu } from '@sourcegraph/client-api'

import type { FlatExtensionHostAPI } from '../api/contract'
import { pretendProxySubscribable, pretendRemote } from '../api/util'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { extensionsController } from '../testing/searchTestHelpers'

import { ActionsNavItems } from './ActionsNavItems'

jest.mock('mdi-react/OpenInNewIcon', () => 'OpenInNewIcon')

describe('ActionItem', () => {
    const NOOP_PLATFORM_CONTEXT = { settings: NEVER }
    const location = H.createLocation(
        'https://github.com/sourcegraph/sourcegraph/pull/5287/files#diff-eb9883bb910397a210512a13fd7384ac'
    )

    test('Renders contributed action items', async () => {
        let component!: RenderResult
        // eslint-disable-next-line @typescript-eslint/require-await
        await act(async () => {
            component = render(
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

        expect(component.asFragment()).toMatchSnapshot()
    })
})
