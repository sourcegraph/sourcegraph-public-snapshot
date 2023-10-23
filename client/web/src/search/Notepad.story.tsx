import React, { useEffect } from 'react'

import type { Meta, StoryFn, StoryObj } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'

import { WebStory } from '../components/WebStory'
import { useNotepadState } from '../stores'
import type { NotepadEntry, NotepadStore } from '../stores/notepad'

import { NotepadContainer } from './Notepad'

const NotepadWrapper: React.FunctionComponent<{ open?: boolean; enableNotepad?: boolean } & NotepadStore> = ({
    entries = [],
    previousEntries = [],
    canRestoreSession = false,
    open = false,
    enableNotepad = true,
}): React.ReactElement => {
    useEffect(() => {
        useNotepadState.setState({ entries, previousEntries, canRestoreSession }, true)
    }, [entries, previousEntries, canRestoreSession])

    return (
        <MockTemporarySettings settings={{ 'search.notepad.enabled': enableNotepad }}>
            <NotepadContainer initialOpen={open} />
        </MockTemporarySettings>
    )
}

const META: Meta<typeof NotepadWrapper> = {
    title: 'web/search/Notepad',
    component: NotepadWrapper,
}
export default META

type Story = StoryObj<typeof META>

const mockEntries: NotepadEntry[] = [
    { id: 0, type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.standard },
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
        patternType: SearchPatternType.standard,
    },
]

const Template: StoryFn<typeof NotepadWrapper> = args => <WebStory>{() => <NotepadWrapper {...args} />}</WebStory>

export const NotepadClosed: Story = Template.bind({})
NotepadClosed.args = {
    entries: mockEntries,
}

export const NotepadClosedEmpty: Story = Template.bind({})
NotepadClosedEmpty.args = {
    entries: [],
}

export const NotepadOpen: Story = Template.bind({})
NotepadOpen.args = {
    entries: mockEntries,
    open: true,
}

export const NotepadRestorePreviousSession: Story = Template.bind({})
NotepadRestorePreviousSession.args = {
    entries: mockEntries,
    open: true,
    canRestoreSession: true,
}

export const NotepadManyEntries: Story = Template.bind({})
NotepadManyEntries.args = {
    entries: Array.from({ length: 50 }, (_element, index) => ({
        id: index,
        type: 'search',
        query: `TODO${index}`,
        caseSensitive: false,
        patternType: SearchPatternType.standard,
    })),
    open: true,
}
