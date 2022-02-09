import { ComponentStory, ComponentMeta } from '@storybook/react'
import React, { useEffect } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../components/WebStory'
import { useExperimentalFeatures, useSearchStackState } from '../stores'
import { SearchStackEntry, SearchStackStore } from '../stores/searchStack'

import { SearchStack } from './SearchStack'

function SearchStackWrapper({
    entries = [],
    previousEntries = [],
    canRestoreSession = false,
    open = false,
    enableSearchStack = true,
}: { open?: boolean; enableSearchStack?: boolean } & SearchStackStore) {
    useExperimentalFeatures.setState({ enableSearchStack })

    useEffect(() => {
        useSearchStackState.setState({ entries, previousEntries, canRestoreSession }, true)
    }, [entries, previousEntries, canRestoreSession])

    return <SearchStack initialOpen={open} />
}

export default {
    title: 'web/search/Search Stack',
    component: SearchStackWrapper,
} as ComponentMeta<typeof SearchStack>

const mockEntries: SearchStackEntry[] = [
    { id: 0, type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.literal },
    { id: 1, type: 'file', path: 'path/to/file1', repo: 'my/repo', revision: 'master', lineRange: null },
    {
        id: 2,
        type: 'file',
        path: 'path/to/a/really/deeply/nested/file/that/should/be/abbreviated/somehow',
        repo: 'github.com/sourcegraph/sourcegraph',
        revision: 'master',
        lineRange: { startLine: 10, endLine: 11 },
    },
    {
        id: 3,
        type: 'search',
        query: 'file:ts$ a really long search query that should wrap',
        caseSensitive: false,
        patternType: SearchPatternType.literal,
    },
]

const Template: ComponentStory<typeof SearchStackWrapper> = args => (
    <WebStory>{() => <SearchStackWrapper {...args} />}</WebStory>
)

export const SearchStackClosed = Template.bind({})
SearchStackClosed.args = {
    entries: mockEntries,
}

export const SearchStackOpen = Template.bind({})
SearchStackOpen.args = {
    entries: mockEntries,
    open: true,
}

export const SearchStackRestorePreviousSession = Template.bind({})
SearchStackRestorePreviousSession.args = {
    entries: mockEntries,
    open: true,
    canRestoreSession: true,
}

export const SearchStackEmptyWithRestore = Template.bind({})
SearchStackEmptyWithRestore.args = {
    entries: [],
    canRestoreSession: true,
}

export const SearchStackEmptyWithoutRestore = Template.bind({})
SearchStackEmptyWithoutRestore.args = {
    entries: [],
    canRestoreSession: false,
}

export const SearchStackManyEntries = Template.bind({})
SearchStackManyEntries.args = {
    entries: Array.from({ length: 50 }, (_element, index) => ({
        id: index,
        type: 'search',
        query: `TODO${index}`,
        caseSensitive: false,
        patternType: SearchPatternType.literal,
    })),
    open: true,
}
