import bitbucketStyles from '@atlassian/aui/dist/aui/css/aui.css'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import browserExtensionStyles from '@sourcegraph/browser/src/app.scss'

import { NotificationType } from '../api/extension/extensionHostApi'
import { registerHighlightContributions } from '../highlight/contributions'

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
}

export default config

registerHighlightContributions()

const BITBUCKET_CLASS_PROPS: HoverOverlayClassProps = {
    className: 'aui-dialog',
    actionItemClassName: 'aui-button hover-action-item--bitbucket-server',
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
)
BitbucketStyles.storyName = 'Bitbucket styles'

export const Branded: Story = () => <BitbucketStyles useBrandedLogo={true} />
