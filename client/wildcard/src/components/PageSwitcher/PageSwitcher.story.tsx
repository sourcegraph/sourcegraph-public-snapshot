import { useState } from 'react'

import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '../../stories/BrandedStory'
import { Text } from '../Typography/Text/Text'

import { PageSwitcher } from './PageSwitcher'

const decorator: Decorator = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageSwitcher',
    component: PageSwitcher,
    decorators: [decorator],
    parameters: {
        component: PageSwitcher,
        design: [
            {
                type: 'figma',
                name: 'Figma',
                url: 'https://www.figma.com/file/LZoW17Fy6eqOfnxjxIRB7d/%F0%9F%93%91-Pagination-Experiments?t=0QPBSel9sN03v8us-1',
            },
        ],
    },
}

export default config

export const Simple: StoryFn = (args = {}) => {
    const totalPages = args.totalCount

    const [page, setPage] = useState(1)

    const goToNextPage = async () => {
        await sleep(2000)
        setPage(page => (page < totalPages ? page + 1 : page))
    }
    const goToPreviousPage = async () => {
        await sleep(2000)
        setPage(page => (page > 1 ? page - 1 : page))
    }
    const goToFirstPage = async () => {
        await sleep(2000)
        setPage(1)
    }
    const goToLastPage = async () => {
        await sleep(2000)
        setPage(totalPages)
    }

    const hasNextPage = page < totalPages
    const hasPreviousPage = page > 1

    return (
        <div>
            <Text alignment="center">
                Showing page {page} of {totalPages}
            </Text>
            <PageSwitcher
                totalLabel={args.totalLabel}
                totalCount={args.totalCount}
                goToNextPage={goToNextPage}
                goToPreviousPage={goToPreviousPage}
                goToFirstPage={goToFirstPage}
                goToLastPage={goToLastPage}
                hasNextPage={hasNextPage}
                hasPreviousPage={hasPreviousPage}
            />
        </div>
    )
}
Simple.argTypes = {
    totalCount: {
        name: 'totalCount',
        control: { type: 'number' },
    },
    totalLabel: {
        name: 'totalLabel',
        control: { type: 'string' },
    },
}
Simple.args = {
    totalCount: 5,
    totalLabel: 'pages',
}

Simple.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

function sleep(timeout: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, timeout))
}
