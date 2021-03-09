import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'
import { number } from '@storybook/addon-knobs'

import webStyles from '../SourcegraphWebApp.scss'
import { PageSelector } from './PageSelector'

const { add } = storiesOf('web/PageSelector', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic', () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} maxPages={number('maxPages', 10)} />
})
