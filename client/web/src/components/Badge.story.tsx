import { select, boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import webStyles from '../SourcegraphWebApp.scss'

import { AppRouterContainer } from './AppRouterContainer'
import { Badge } from './Badge'

const { add } = storiesOf('web/Badge', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <AppRouterContainer>
            <div className="container mt-3">{story()}</div>
        </AppRouterContainer>
    </>
))

add('Basic', () => {
    const status = select('Status', { Beta: 'beta', Prototype: 'prototype' }, 'beta')
    const useLink = boolean('Use link', false)

    return <Badge status={status} tooltip="This is a tooltip" useLink={useLink} />
})
