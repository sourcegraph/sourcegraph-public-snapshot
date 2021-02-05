import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'
import { GrowingTextArea } from './GrowingTextArea'

const { add } = storiesOf('web/GrowingTextArea', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => <GrowingTextArea placeholder="Beans" />)
