import * as H from 'history'
import React from 'react'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { ActionsNavItems } from './ActionsNavItems'
import { ContributableMenu } from '../api/protocol'
import { of, NEVER } from 'rxjs'
import { Services } from '../api/client/services'
import { mount } from 'enzyme'

jest.mock('mdi-react/OpenInNewIcon', () => 'OpenInNewIcon')

describe('ActionItem', () => {
    const MOCK_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve(undefined) }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined, settings: NEVER }
    const location = H.createLocation(
        'https://github.com/sourcegraph/sourcegraph/pull/5287/files#diff-eb9883bb910397a210512a13fd7384ac'
    )

    test('Renders contributed action items', () => {
        const component = mount(
            <ActionsNavItems
                menu={ContributableMenu.EditorTitle}
                location={location}
                extensionsController={{
                    ...MOCK_EXTENSIONS_CONTROLLER,
                    services: {
                        contribution: {
                            getContributions: () =>
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
                                }),
                        },
                    } as Services,
                }}
                platformContext={NOOP_PLATFORM_CONTEXT}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })
})
