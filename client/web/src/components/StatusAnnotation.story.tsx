import React from 'react'
import { storiesOf } from '@storybook/react'
import { select } from '@storybook/addon-knobs'
import webStyles from '../SourcegraphWebApp.scss'
import { StatusAnnotation } from './StatusAnnotation'

const { add } = storiesOf('web/StatusAnnotation', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => {
    const status = select('Status', { Beta: 'beta', Prototype: 'prototype' }, 'beta')
    return <StatusAnnotation status={status} />
})
