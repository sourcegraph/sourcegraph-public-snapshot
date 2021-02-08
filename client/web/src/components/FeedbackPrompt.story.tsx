import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'
import { FeedbackPrompt } from './FeedbackPrompt'

const { add } = storiesOf('web/FeedbackPrompt', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => <FeedbackPrompt placeholder="Beans" />)
