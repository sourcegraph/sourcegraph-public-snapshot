import { useEffect, useRef } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subMonths } from 'date-fns'
import { useDarkMode } from 'storybook-dark-mode'

import { SymbolKind } from '@sourcegraph/search'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { usePrependStyles } from '@sourcegraph/storybook'

import { applyTheme } from '..'
import { dark } from '../../bridge-mock/theme-snapshots/dark'
import { light } from '../../bridge-mock/theme-snapshots/light'

import { SearchResultList } from './SearchResultList'

import globalStyles from '../../index.scss'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'jetbrains/SearchResultList',
    decorators: [decorator],
}

export default config

// Use a consistent diff for date to avoid monthly snapshot failures
const AUTHOR_DATE = subMonths(new Date(), 7).toISOString()

export const JetBrainsSearchResultListStory: Story = () => {
    const rootElementRef = useRef<HTMLDivElement>(null)
    const isDarkTheme = useDarkMode()

    const matches: SearchMatch[] = [
        {
            type: 'content',
            path: '/CHANGELOG.md',
            repository: 'test-repository',
            repoStars: 1,
            branches: ['a', 'b'],
            commit: 'hunk12ef',
            lineMatches: [
                { line: 'Test line 1', lineNumber: 0, offsetAndLengths: [] },
                { line: 'Test line 5', lineNumber: 4, offsetAndLengths: [] },
            ],
        },
        {
            type: 'repo',
            repository: 'test-repository',
            repoStars: 2,
            description: 'Repo description',
            fork: true,
            archived: true,
            private: true,
            branches: ['a', 'b'],
        },
        {
            type: 'commit',
            url: 'https://github.com/sourcegraph/sourcegraph',
            repository: 'test-repository',
            oid: 'hunk12ef',
            message: 'Commit message',
            authorName: 'Test User',
            authorDate: AUTHOR_DATE,
            repoStars: 3,
            content: '',
            // Array of [line, character, length] triplets
            ranges: [],
        },
        {
            type: 'symbol',
            path: '/CHANGELOG.md',
            repository: 'test-repository',
            repoStars: 4,
            branches: ['a', 'b'],
            commit: 'hunk12ef',
            symbols: [
                {
                    url: 'https://github.com/sourcegraph/sourcegraph',
                    name: 'TestSymbol',
                    containerName: 'TestContainer',
                    kind: SymbolKind.CONSTANT,
                    line: 1,
                },
            ],
        },
        {
            type: 'path',
            path: '/CHANGELOG.md',
            repository: 'test-repository',
            repoStars: 5,
            branches: ['a', 'b'],
            commit: 'hunk12ef',
        },
    ]

    usePrependStyles('branded-story-styles', globalStyles)

    useEffect(() => {
        if (rootElementRef.current === null) {
            return
        }
        applyTheme(isDarkTheme ? dark : light, rootElementRef.current)
    }, [rootElementRef, isDarkTheme])

    return (
        <div ref={rootElementRef}>
            <div className="d-flex justify-content-center">
                <div className="mx-6">
                    <SearchResultList
                        matches={matches}
                        onPreviewChange={async () => {}}
                        onPreviewClear={async () => {}}
                        onOpen={async () => {}}
                    />
                </div>
            </div>
        </div>
    )
}

JetBrainsSearchResultListStory.parameters = {
    chromatic: {
        disableSnapshot: false,
    },
}
