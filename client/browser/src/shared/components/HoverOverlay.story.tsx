import type { Decorator, Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'
import { BrowserRouter } from 'react-router-dom'

import { registerHighlightContributions } from '@sourcegraph/common'
import { HoverOverlay, type HoverOverlayClassProps } from '@sourcegraph/shared/src/hover/HoverOverlay'
import {
    commonProps,
    FIXTURE_ACTIONS,
    FIXTURE_CONTENT,
    FIXTURE_SEMANTIC_BADGE,
} from '@sourcegraph/shared/src/hover/HoverOverlay.fixtures'

import browserExtensionStyles from '../../app.scss'
import bitbucketCodeHostStyles from '../code-hosts/bitbucket/codeHost.module.scss'
import bitbucketStyles from '@atlassian/aui/dist/aui/css/aui.css'

const decorator: Decorator = story => (
    <>
        <style>{bitbucketStyles}</style>
        <style>{browserExtensionStyles}</style>
        {story()}
    </>
)

const config: Meta = {
    // This story lives in the @sourcegraph/browser package because
    // it uses the browser extension styles and bitbucket CSS module styles.
    title: 'shared/HoverOverlay',
    decorators: [decorator],
    parameters: {},
}

export default config

registerHighlightContributions()

const BITBUCKET_CLASS_PROPS: HoverOverlayClassProps = {
    className: 'aui-dialog',
    actionItemClassName: classNames('aui-button', bitbucketCodeHostStyles.hoverActionItem),
    iconClassName: 'aui-icon',
}

export const BitbucketStyles: StoryFn = (props = {}) => (
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

export const Branded: StoryFn = () => <BitbucketStyles useBrandedLogo={true} />
