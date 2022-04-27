import bitbucketStyles from '@atlassian/aui/dist/aui/css/aui.css'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'
import { BrowserRouter } from 'react-router-dom'

// eslint-disable-next-line no-restricted-imports
import browserExtensionStyles from '@sourcegraph/browser/src/app.scss'
// eslint-disable-next-line no-restricted-imports
import bitbucketCodeHostStyles from '@sourcegraph/browser/src/shared/code-hosts/bitbucket/codeHost.module.scss'
import { registerHighlightContributions } from '@sourcegraph/common'

import { NotificationType } from '../api/extension/extensionHostApi'

import { HoverOverlay, HoverOverlayClassProps } from './HoverOverlay'
import { commonProps, FIXTURE_ACTIONS, FIXTURE_CONTENT, FIXTURE_SEMANTIC_BADGE } from './HoverOverlay.fixtures'

const decorator: DecoratorFn = story => (
    <>
        <style>{bitbucketStyles}</style>
        <style>{browserExtensionStyles}</style>
        {story()}
    </>
)

const config: Meta = {
    title: 'shared/HoverOverlay',
    decorators: [decorator],
    parameters: {
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

registerHighlightContributions()

const BITBUCKET_CLASS_PROPS: HoverOverlayClassProps = {
    className: 'aui-dialog',
    actionItemClassName: classNames('aui-button', bitbucketCodeHostStyles.hoverActionItem),
    iconClassName: 'aui-icon',
    getAlertClassName: alertKind => {
        switch (alertKind) {
            case NotificationType.Error:
                return 'aui-message aui-message-error'
            default:
                return 'aui-message aui-message-info'
        }
    },
}

export const BitbucketStyles: Story = props => (
    <BrowserRouter>
        <HoverOverlay
            {...commonProps()}
            {...BITBUCKET_CLASS_PROPS}
            {...props}
            hoverOrError={{
                contents: [FIXTURE_CONTENT],
                aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
            }}
            actionsOrError={FIXTURE_ACTIONS}
        />
    </BrowserRouter>
)
BitbucketStyles.storyName = 'Bitbucket styles'

export const Branded: Story = () => <BitbucketStyles useBrandedLogo={true} />
