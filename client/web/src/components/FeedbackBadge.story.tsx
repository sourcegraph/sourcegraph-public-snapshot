import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'

import { AppRouterContainer } from './AppRouterContainer'
import { FeedbackBadge } from './FeedbackBadge'

const { add } = storiesOf('web/Badge', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <AppRouterContainer>
            <div className="container mt-3">{story()}</div>
        </AppRouterContainer>
    </>
))

add('Feedback', () => {
    const status = select('Status', { Beta: 'beta', Prototype: 'prototype' }, 'beta')
    return (
        <FeedbackBadge status={status} tooltip="This is a tooltip" feedback={{ mailto: 'support@sourcegraph.com' }} />
    )
})
