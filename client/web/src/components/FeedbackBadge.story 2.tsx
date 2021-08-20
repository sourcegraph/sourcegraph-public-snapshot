import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'

import { FeedbackBadge } from './FeedbackBadge'

const { add } = storiesOf('web/Badge', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container mt-3">{story()}</div>
        </div>
    </>
))

add('Feedback', () => {
    const status = select('Status', { Beta: 'beta', Prototype: 'prototype' }, 'beta')
    return (
        <FeedbackBadge status={status} tooltip="This is a tooltip" feedback={{ mailto: 'support@sourcegraph.com' }} />
    )
})
