import bitbucketStyles from '@atlassian/aui/dist/aui/css/aui.css'
import { action } from '@storybook/addon-actions'
import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import React from 'react'
import { of } from 'rxjs'
import { MarkupContent, Badged, AggregableBadge } from 'sourcegraph'

import browserExtensionStyles from '@sourcegraph/browser/src/app.scss'
import { MarkupKind } from '@sourcegraph/extension-api-classes'

import { registerHighlightContributions } from '../highlight/contributions'
import { PlatformContext } from '../platform/context'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'

import { HoverOverlay, HoverOverlayClassProps } from './HoverOverlay'

registerHighlightContributions()

const { add } = storiesOf('shared/HoverOverlay', module)

const history = createMemoryHistory()
const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve() }
const NOOP_PLATFORM_CONTEXT: Pick<PlatformContext, 'forceUpdateTooltip' | 'settings'> = {
    forceUpdateTooltip: () => undefined,
    settings: of({ final: {}, subjects: [] }),
}

export const commonProps = () => ({
    showCloseButton: boolean('showCloseButton', true),
    location: history.location,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    extensionsController: NOOP_EXTENSIONS_CONTROLLER,
    platformContext: NOOP_PLATFORM_CONTEXT,
    isLightTheme: true,
    overlayPosition: { top: 16, left: 16 },
    onAlertDismissed: action('onAlertDismissed'),
    onCloseButtonClick: action('onCloseButtonClick'),
})

const bitbucketClassProps: HoverOverlayClassProps = {
    className: 'aui-dialog',
    actionItemClassName: 'aui-button hover-action-item--bitbucket-server',
    iconButtonClassName: 'aui-button btn-icon--bitbucket-server',
    infoAlertClassName: 'aui-message aui-message-info',
    errorAlertClassName: 'aui-message aui-message-error',
    iconClassName: 'aui-icon',
}

export const FIXTURE_CONTENT: Badged<MarkupContent> = {
    value:
        '```go\nfunc RegisterMiddlewares(m ...*Middleware)\n```\n\n' +
        '---\n\nRegisterMiddlewares registers additional authentication middlewares. Currently this is used to register enterprise-only SSO middleware. This should only be called from an init function.\n',
    kind: MarkupKind.Markdown,
}

export const FIXTURE_SEMANTIC_BADGE: AggregableBadge = {
    text: 'semantic',
    linkURL: 'https://docs.sourcegraph.com/code_intelligence/explanations/precise_code_intelligence',
    hoverMessage: 'Sample hover message',
}

export const FIXTURE_ACTIONS = [
    {
        action: {
            id: 'goToDefinition.preloaded',
            title: 'Go to definition',
            command: 'open',
            commandArguments: ['/github.com/sourcegraph/codeintellify/-/blob/src/hoverifier.ts?subtree=true#L57:1'],
        },
        active: true,
    },
    {
        action: {
            id: 'findReferences',
            title: 'Find references',
            command: 'open',
            commandArguments: [
                '/github.com/sourcegraph/codeintellify/-/blob/src/hoverifier.ts?subtree=true#L57:18&tab=references',
            ],
        },
        active: true,
    },
]

add('Bitbucket styles', () => (
    <>
        <style>{bitbucketStyles}</style>
        <style>{browserExtensionStyles}</style>
        <HoverOverlay
            {...commonProps()}
            {...bitbucketClassProps}
            hoverOrError={{
                contents: [FIXTURE_CONTENT],
                aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
            }}
            actionsOrError={FIXTURE_ACTIONS}
        />
    </>
))
