import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'
import { number } from '@storybook/addon-knobs'
import webStyles from '../../../../web/src/SourcegraphWebApp.scss'
import { PageSelector } from './PageSelector'
import { BrandedStory } from '../../../../branded/src/components/BrandedStory'

const { add } = storiesOf('wildcard/PageSelector', module).addDecorator(story => (
    <BrandedStory styles={webStyles}>{() => <div className="container web-content mt-3">{story()}</div>}</BrandedStory>
))

add('Short typical', () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 5)} />
})

add('Long typical', () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
})

add('Long active', () => {
    const [page, setPage] = useState(5)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
})

add('Long complete', () => {
    const [page, setPage] = useState(10)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
})
