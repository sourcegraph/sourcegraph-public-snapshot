import { Meta, Story } from '@storybook/react'
import { createBrowserHistory } from 'history'
import { noop } from 'lodash'
import React from 'react'
import { of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { FlatExtensionHostAPI } from '../api/contract'
import { pretendProxySubscribable, pretendRemote } from '../api/util'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { NOOP_SETTINGS_CASCADE } from '../util/searchTestHelpers'

import { CommandPalette, CommandPaletteProps } from './CommandPalette'

const config: Meta = {
    title: 'shared/CommandPallette',

    parameters: {
        component: CommandPalette,
    },
}

export default config

const commandPaletteActions = [
    {
        id: 'a',
        actionItem: {
            label: 'Action A',
            description: 'This is Action A',
        },
        command: 'open',
        commandArguments: ['https://example.com'],
    },
    {
        id: 'b',
        actionItem: {
            label: 'Action B',
            description: 'This is Action B',
        },
        command: 'updateConfiguration',
        commandArguments: [],
    },
]

const commandPaletteMenus = {
    commandPalette: [
        {
            action: 'a',
        },
        {
            action: 'b',
        },
    ],
}

export const Default: Story<CommandPaletteProps> = () => {
    const history = createBrowserHistory()

    return (
        <BrandedStory styles={webStyles}>
            {() => (
                <CommandPalette
                    isOpen={true}
                    extensionsController={{
                        extHostAPI: Promise.resolve(
                            pretendRemote<FlatExtensionHostAPI>({
                                haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                                getContributions: () =>
                                    pretendProxySubscribable(
                                        of({ actions: commandPaletteActions, menus: commandPaletteMenus })
                                    ),
                            })
                        ),
                        executeCommand: (commandID, commandArguments) => {
                            console.log({ commandID, commandArgs: commandArguments })
                            return Promise.resolve()
                        },
                    }}
                    platformContext={{ forceUpdateTooltip: noop, settings: of(NOOP_SETTINGS_CASCADE) }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    location={history.location}
                />
            )}
        </BrandedStory>
    )
}
