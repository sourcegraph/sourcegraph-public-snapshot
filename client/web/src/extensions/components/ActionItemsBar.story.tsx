import React from 'react'

import { DecoratorFn, Meta } from '@storybook/react'
import * as H from 'history'
import { EMPTY, noop, of } from 'rxjs'

import { ContributableMenu, Contributions, Evaluated } from '@sourcegraph/client-api'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { extensionsController, NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { AppRouterContainer } from '../../components/AppRouterContainer'
import { WebStory } from '../../components/WebStory'
import { SourcegraphContext } from '../../jscontext'

import { ActionItemsBar, useWebActionItems } from './ActionItemsBar'

import webStyles from '../../SourcegraphWebApp.scss'

const LOCATION: H.Location = {
    search: '',
    hash: '',
    pathname: '/github.com/sourcegraph/sourcegraph/-/blob/client/browser/src/browser-extension/ThemeWrapper.tsx',
    key: 'oq2z4k',
    state: undefined,
}

const mockIconURL =
    'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDE2LjAuMCwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4iICJodHRwOi8vd3d3LnczLm9yZy9HcmFwaGljcy9TVkcvMS4xL0RURC9zdmcxMS5kdGQiPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IkxheWVyXzEiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiIHg9IjBweCIgeT0iMHB4IgoJIHdpZHRoPSI5N3B4IiBoZWlnaHQ9Ijk3cHgiIHZpZXdCb3g9IjAgMCA5NyA5NyIgZW5hYmxlLWJhY2tncm91bmQ9Im5ldyAwIDAgOTcgOTciIHhtbDpzcGFjZT0icHJlc2VydmUiPgo8Zz4KCTxwYXRoIGZpbGw9IiNGMDUxMzMiIGQ9Ik05Mi43MSw0NC40MDhMNTIuNTkxLDQuMjkxYy0yLjMxLTIuMzExLTYuMDU3LTIuMzExLTguMzY5LDBsLTguMzMsOC4zMzJMNDYuNDU5LDIzLjE5CgkJYzIuNDU2LTAuODMsNS4yNzItMC4yNzMsNy4yMjksMS42ODVjMS45NjksMS45NywyLjUyMSw0LjgxLDEuNjcsNy4yNzVsMTAuMTg2LDEwLjE4NWMyLjQ2NS0wLjg1LDUuMzA3LTAuMyw3LjI3NSwxLjY3MQoJCWMyLjc1LDIuNzUsMi43NSw3LjIwNiwwLDkuOTU4Yy0yLjc1MiwyLjc1MS03LjIwOCwyLjc1MS05Ljk2MSwwYy0yLjA2OC0yLjA3LTIuNTgtNS4xMS0xLjUzMS03LjY1OGwtOS41LTkuNDk5djI0Ljk5NwoJCWMwLjY3LDAuMzMyLDEuMzAzLDAuNzc0LDEuODYxLDEuMzMyYzIuNzUsMi43NSwyLjc1LDcuMjA2LDAsOS45NTljLTIuNzUsMi43NDktNy4yMDksMi43NDktOS45NTcsMGMtMi43NS0yLjc1NC0yLjc1LTcuMjEsMC05Ljk1OQoJCWMwLjY4LTAuNjc5LDEuNDY3LTEuMTkzLDIuMzA3LTEuNTM3VjM2LjM2OWMtMC44NC0wLjM0NC0xLjYyNS0wLjg1My0yLjMwNy0xLjUzN2MtMi4wODMtMi4wODItMi41ODQtNS4xNC0xLjUxNi03LjY5OAoJCUwzMS43OTgsMTYuNzE1TDQuMjg4LDQ0LjIyMmMtMi4zMTEsMi4zMTMtMi4zMTEsNi4wNiwwLDguMzcxbDQwLjEyMSw0MC4xMThjMi4zMSwyLjMxMSw2LjA1NiwyLjMxMSw4LjM2OSwwTDkyLjcxLDUyLjc3OQoJCUM5NS4wMjEsNTAuNDY4LDk1LjAyMSw0Ni43MTksOTIuNzEsNDQuNDA4eiIvPgo8L2c+Cjwvc3ZnPgo='

if (!window.context) {
    window.context = { enableLegacyExtensions: false } as SourcegraphContext & Mocha.SuiteFunction
}

// eslint-disable-next-line id-length
const mockActionItems = [...(new Array(10) as (number | undefined)[])].map((_, index) => `${index}`)
const mockContributions: Evaluated<Contributions> = {
    actions: mockActionItems.map((id, index) => ({
        id,
        actionItem: {
            iconURL: mockIconURL,
            label: 'Some label',
            pressed: index < 2,
            description: 'Some description',
        },
        command: 'open',
        active: true,
        category: 'Some category',
        title: 'Some title',
    })),
    menus: {
        [ContributableMenu.EditorTitle]: mockActionItems.map(id => ({
            action: id,
            alt: id,
            when: true,
        })),
    },
}

const mockExtensionsController = {
    ...extensionsController,
    extHostAPI: Promise.resolve(
        pretendRemote<FlatExtensionHostAPI>({
            getContributions: () => pretendProxySubscribable(of(mockContributions)),
            registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
            haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
        })
    ),
}

const decorator: DecoratorFn = story => (
    <>
        <style>{webStyles}</style>
        <WebStory>
            {() => (
                <AppRouterContainer>
                    <div className="container mt-3">{story()}</div>
                </AppRouterContainer>
            )}
        </WebStory>
    </>
)

const config: Meta = {
    title: 'web/extensions/ActionItemsBar',
    decorators: [decorator],
    component: ActionItemsBar,
    parameters: {
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const Default: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const { useActionItemsBar } = useWebActionItems()

    return (
        <ActionItemsBar
            repo={undefined}
            location={LOCATION}
            useActionItemsBar={useActionItemsBar}
            extensionsController={mockExtensionsController}
            platformContext={NOOP_PLATFORM_CONTEXT as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
        />
    )
}
