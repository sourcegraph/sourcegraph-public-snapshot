import { number } from '@storybook/addon-knobs'
import { Meta } from '@storybook/react'
import React, { useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { PageSelector } from './PageSelector'

const Story: Meta = {
    title: 'wildcard/PageSelector',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: PageSelector,
    },
}

export default Story

export const Short = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 5)} />
}

export const Long = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}

export const LongOnMobile = () => {
    const [page, setPage] = useState(1)
    return (
        <div style={{ width: 320 }}>
            <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
        </div>
    )
}

LongOnMobile.story = {
    name: 'Long on mobile',
}

export const LongActive = () => {
    const [page, setPage] = useState(5)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}

LongActive.story = {
    name: 'Long active',
}

export const LongComplete = () => {
    const [page, setPage] = useState(10)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}

LongComplete.story = {
    name: 'Long complete',
}
