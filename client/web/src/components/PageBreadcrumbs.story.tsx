import { storiesOf } from '@storybook/react'
import React from 'react'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import webStyles from '../SourcegraphWebApp.scss'
import { PageBreadcrumbs } from './PageBreadcrumbs'

const { add } = storiesOf('web/PageBreadcrumbs', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => (
    <PageBreadcrumbs icon={PuzzleOutlineIcon} path={[{ to: '/level-1', text: 'Level 1' }, { text: 'Level 2' }]} />
))
