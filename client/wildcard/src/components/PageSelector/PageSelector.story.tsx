import { useState } from 'react'

import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { H1, H2 } from '..'

import { PageSelector } from './PageSelector'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageSelector',
    component: PageSelector,
    decorators: [decorator],
}

export default config

export const Simple: Story = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={number('maxPages', 5)} />
}

export const AllPageSelectors: Story = () => (
    <>
        <H1>Page Selector</H1>
        <H2>Short</H2>
        <Short />
        <H2>Long</H2>
        <Long />
        <H2>Long active</H2>
        <LongActive />
        <H2>Long complete</H2>
        <LongComplete />
        <H2>Long on mobile</H2>
        <LongOnMobile />
    </>
)

AllPageSelectors.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

const Short = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={5} />
}

const Long = () => {
    const [page, setPage] = useState(1)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}

const LongOnMobile = () => {
    const [page, setPage] = useState(1)
    return (
        <div style={{ width: 320 }}>
            <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
        </div>
    )
}

const LongActive = () => {
    const [page, setPage] = useState(5)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}

const LongComplete = () => {
    const [page, setPage] = useState(10)
    return <PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
}
