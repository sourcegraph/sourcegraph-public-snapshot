import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { PageSelector } from './PageSelector'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageSelector',
    decorators: [decorator],
}

export default config

export const Short: Story = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 5)} />
}

export const Long: Story = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
}

export const LongOnMobile: Story = () => {
    const [page, setPage] = useState(1)
    return (
        <div style={{ width: 320 }}>
            <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
        </div>
    )
}

LongOnMobile.storyName = 'Long on mobile'

export const LongActive: Story = () => {
    const [page, setPage] = useState(5)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
}

LongActive.storyName = 'Long active'

export const LongComplete: Story = () => {
    const [page, setPage] = useState(10)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 10)} />
}

LongComplete.storyName = 'Long complete'
