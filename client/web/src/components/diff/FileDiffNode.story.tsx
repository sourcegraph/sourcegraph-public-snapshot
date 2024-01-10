import type { Decorator, Meta, StoryFn } from '@storybook/react'

import type { FileDiffFields } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { DEMO_HUNKS } from './FileDiffHunks.story'
import { FileDiffNode } from './FileDiffNode'

export const FILE_DIFF_NODES: FileDiffFields[] = [
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        oldFile: null,
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 2,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/new_file.md',
            changelistURL: null,
        },
        newPath: 'new_file.md',
        oldPath: null,
    },
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        newFile: null,
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 2,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/deleted_file.md',
            changelistURL: null,
        },
        newPath: null,
        oldPath: 'deleted_file.md',
    },
    {
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, deleted: 0 },
        oldFile: null,
        newFile: {
            __typename: 'VirtualFile',
            binary: true,
            byteSize: 1280,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/new_file.md',
            changelistURL: null,
        },
        newPath: 'new_file.md',
        oldPath: null,
    },
    {
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, deleted: 0 },
        newFile: null,
        oldFile: {
            __typename: 'VirtualFile',
            binary: true,
            byteSize: 1280,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/deleted_file.md',
            changelistURL: null,
        },
        newPath: null,
        oldPath: 'deleted_file.md',
    },
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/existing_file.md',
            changelistURL: null,
        },
        newPath: 'existing_file.md',
        oldPath: 'existing_file.md',
    },
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        oldFile: {
            __typename: 'GitBlob',
            binary: false,
            byteSize: 0,
        },
        newFile: {
            __typename: 'GitBlob',
            binary: false,
            byteSize: 0,
        },
        mostRelevantFile: {
            __typename: 'GitBlob',
            url: 'http://test.test/gitblob',
            changelistURL: null,
        },
        newPath: 'existing_git_file.md',
        oldPath: 'existing_git_file.md',
    },
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/to.md',
            changelistURL: null,
        },
        newPath: 'to.md',
        oldPath: 'from.md',
    },
    {
        hunks: DEMO_HUNKS,
        internalID: 'abcdef123',
        stat: { added: 1, deleted: 1 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: 'dir2/to.md',
            changelistURL: null,
        },
        newPath: 'dir2/to.md',
        oldPath: 'dir1/from.md',
    },
    {
        hunks: [],
        internalID: 'abcdef123',
        stat: { added: 0, deleted: 0 },
        oldFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        newFile: {
            __typename: 'VirtualFile',
            binary: false,
            byteSize: 0,
        },
        mostRelevantFile: {
            __typename: 'VirtualFile',
            url: '/to.md',
            changelistURL: null,
        },
        newPath: 'to.md',
        oldPath: 'from.md',
    },
]

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/diffs/FileDiffNode',
    decorators: [decorator],
    includeStories: ['AllUnifiedFileNode', 'AllSplitFileNode'],
    argTypes: {
        persistLines: {
            control: { type: 'boolean' },
        },
        lineNumbers: {
            control: { type: 'boolean' },
        },
    },
    args: {
        persistLines: true,
        lineNumbers: true,
    },
}

export default config

export const AllUnifiedFileNode: StoryFn = args => (
    <WebStory>
        {webProps => (
            <ul className="list-unstyled">
                {FILE_DIFF_NODES.map((node, index) => (
                    <FileDiffNode
                        {...webProps}
                        diffMode="unified"
                        key={index}
                        persistLines={args.persistLines}
                        lineNumbers={args.lineNumbers}
                        node={node}
                        className="abcdef"
                    />
                ))}
            </ul>
        )}
    </WebStory>
)

AllUnifiedFileNode.storyName = 'All unified file node states overview'

export const AllSplitFileNode: StoryFn = args => (
    <WebStory>
        {webProps => (
            <ul className="list-unstyled">
                {FILE_DIFF_NODES.map((node, index) => (
                    <FileDiffNode
                        {...webProps}
                        diffMode="split"
                        key={index}
                        persistLines={args.persistLines}
                        lineNumbers={args.lineNumbers}
                        node={node}
                        className="abcdef"
                    />
                ))}
            </ul>
        )}
    </WebStory>
)

AllSplitFileNode.storyName = 'All split file node states overview'
