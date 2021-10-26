import bitbucketStyles from '@atlassian/aui/dist/aui/css/aui.css'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import browserExtensionStyles from '@sourcegraph/browser/src/app.scss'

import { NotificationType } from '../api/extension/extensionHostApi'
import { registerHighlightContributions } from '../highlight/contributions'

import { HoverOverlay, HoverOverlayClassProps } from './HoverOverlay'
import { commonProps, FIXTURE_ACTIONS, FIXTURE_CONTENT, FIXTURE_SEMANTIC_BADGE } from './HoverOverlay.fixtures'

registerHighlightContributions()

const bitbucketClassProps: HoverOverlayClassProps = {
    className: 'aui-dialog',
    actionItemClassName: 'aui-button hover-action-item--bitbucket-server',
    closeButtonClassName: 'aui-button btn-icon--bitbucket-server close',
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

const config: Meta = {
    title: 'shared/HoverOverlay',
}
export default config

export const BitbucketStyles: Story = () => (
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
)

BitbucketStyles.storyName = 'Bitbucket styles'
