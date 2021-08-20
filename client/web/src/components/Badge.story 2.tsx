import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'

import { Badge } from './Badge'

const { add } = storiesOf('web/Badge', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => {
    const status = select('Status', { Beta: 'beta', Prototype: 'prototype' }, 'beta')

    return <Badge status={status} tooltip="This is a tooltip" />
})
