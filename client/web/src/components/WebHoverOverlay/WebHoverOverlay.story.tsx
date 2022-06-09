import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'

import { registerHighlightContributions } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import {
    commonProps,
    FIXTURE_ACTIONS,
    FIXTURE_SEMANTIC_BADGE,
    FIXTURE_CONTENT,
} from '@sourcegraph/shared/src/hover/HoverOverlay.fixtures'

import { WebStory } from '../WebStory'

import { WebHoverOverlay } from './WebHoverOverlay'
import {
    FIXTURE_CONTENT_LONG_CODE,
    FIXTURE_CONTENT_LONG_TEXT_ONLY,
    FIXTURE_CONTENT_MARKDOWN,
    FIXTURE_PARTIAL_BADGE,
    FIXTURE_SMALL_TEXT_MARKDOWN_ALERT,
    FIXTURE_WARNING_MARKDOWN_ALERT,
} from './WebHoverOverlay.fixtures'

registerHighlightContributions()

const { add } = storiesOf('web/WebHoverOverlay', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=2877%3A35469',
        },
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    })

add('Loading', () => <WebHoverOverlay {...commonProps()} hoverOrError="loading" actionsOrError={FIXTURE_ACTIONS} />)

add('Error', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={
            new Error(
                'Something terrible happened: Eiusmod voluptate deserunt in sint cillum pariatur laborum eiusmod.'
            )
        }
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('No hover information', () => (
    <WebHoverOverlay {...commonProps()} hoverOrError={null} actionsOrError={FIXTURE_ACTIONS} />
))

add('Common content without actions', () => (
    <WebHoverOverlay {...commonProps()} hoverOrError={{ contents: [FIXTURE_CONTENT] }} />
))

add('Common content with actions', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Aggregated Badges', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Long code', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT_LONG_CODE],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Long text only', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT_LONG_TEXT_ONLY],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Long markdown with <div>', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT_MARKDOWN],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Multiple MarkupContents', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT, FIXTURE_CONTENT, FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('With small-text alert', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [FIXTURE_SMALL_TEXT_MARKDOWN_ALERT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With one-line alert', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.PlainText,
                        value: 'This is a test alert.',
                    },
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With alert with warning icon', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.PlainText,
                        value: 'This is a warning alert.',
                    },
                    iconKind: 'warning',
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With dismissible alert with icon', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.Markdown,
                        value:
                            'Search based result.<br /> [Learn more about precise code intelligence](https://sourcegraph.com/github.com/sourcegraph/code-intel-extensions/-/blob/shared/indicators.ts#L67)',
                    },
                    type: 'test-alert-type',
                    iconKind: 'info',
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With long markdown text and dismissible alert with icon.', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [FIXTURE_WARNING_MARKDOWN_ALERT],
            aggregatedBadges: [FIXTURE_PARTIAL_BADGE, FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('Multiple MarkupContents with badges and alerts', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT, FIXTURE_CONTENT, FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
            alerts: [FIXTURE_SMALL_TEXT_MARKDOWN_ALERT, FIXTURE_WARNING_MARKDOWN_ALERT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))
