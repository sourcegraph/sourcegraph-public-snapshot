import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { FileDiffNode } from './FileDiffNode'
import { createMemoryHistory } from 'history'
import webStyles from '../../SourcegraphWebApp.scss'
import { DEMO_HUNKS } from './FileDiffHunks.story'
import { MemoryRouter } from 'react-router'
import { GitBlob } from '../../../../shared/src/graphql/schema'
import { FileDiffFields } from '../../graphql-operations'

export const FILE_DIFF_NODES: FileDiffFields[] = [
    {
        __typename: 'FileDiff',
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 0, changed: 1, deleted: 0 },
        oldFile: null,
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 2,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/new_file.md',
        },
        newPath: 'new_file.md',
        oldPath: null,
    },
    {
        __typename: 'FileDiff',
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 0, changed: 1, deleted: 0 },
        newFile: null,
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 2,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'deleted_file.md',
        },
        newPath: null,
        oldPath: 'deleted_file.md',
    },
    {
        __typename: 'FileDiff',
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, changed: 0, deleted: 0 },
        oldFile: null,
        newFile: {
            __typename: 'VirtualFile',
            binary: true,
            name: 'new_file.md',
            byteSize: 1280,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: true,
            name: 'new_file.md',
            byteSize: 1280,
        },
        newPath: 'new_file.md',
        oldPath: null,
    },
    {
        __typename: 'FileDiff',
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, changed: 0, deleted: 0 },
        newFile: null,
        oldFile: {
            __typename: 'VirtualFile',
            binary: true,
            name: 'deleted_file.md',
            byteSize: 1280,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: true,
            name: 'deleted_file.md',
            byteSize: 1280,
        },
        newPath: null,
        oldPath: 'deleted_file.md',
    },
    {
        __typename: 'FileDiff',
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 0, changed: 1, deleted: 0 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'existing_file.md',
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'existing_file.md',
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'existing_file.md',
        },
        newPath: 'existing_file.md',
        oldPath: 'existing_file.md',
    },
    {
        __typename: 'FileDiff',
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 0, changed: 1, deleted: 0 },
        oldFile: {
            __typename: 'GitBlob',
            binary: false,
            name: 'existing_git_file.md',
        } as GitBlob,
        newFile: {
            __typename: 'GitBlob',
            binary: false,
            name: 'existing_git_file.md',
        } as GitBlob,
        mostRelevantFile: {
            __typename: 'GitBlob',
            binary: false,
            name: 'existing_git_file.md',
        } as GitBlob,
        newPath: 'existing_git_file.md',
        oldPath: 'existing_git_file.md',
    },
    {
        __typename: 'FileDiff',
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 0, changed: 1, deleted: 0 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'from.md',
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'to.md',
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'to.md',
        },
        newPath: 'to.md',
        oldPath: 'from.md',
    },
    {
        __typename: 'FileDiff',
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, changed: 0, deleted: 0 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'from.md',
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'to.md',
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            binary: false,
            name: 'to.md',
        },
        newPath: 'to.md',
        oldPath: 'from.md',
    },
]

const { add } = storiesOf('web/FileDiffNode', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </>
    )
})

add('All file node states overview', () => (
    <MemoryRouter>
        {FILE_DIFF_NODES.map((node, index) => (
            <FileDiffNode
                key={index}
                persistLines={boolean('persistLines', false)}
                lineNumbers={boolean('lineNumbers', true)}
                isLightTheme={true}
                node={node}
                className="abcdef"
                location={createMemoryHistory().location}
                history={createMemoryHistory()}
            />
        ))}
    </MemoryRouter>
))
