import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { type FileDiffHunkFields, DiffHunkLineType } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { FileDiffHunks } from './FileDiffHunks'

export const DEMO_HUNKS: FileDiffHunkFields[] = [
    {
        oldRange: { lines: 7, startLine: 3 },
        newRange: { lines: 7, startLine: 3 },
        oldNoNewlineAt: false,
        section: 'func awesomeness(param string) (int, error) {',
        highlight: {
            aborted: false,
            lines: [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    v, err := makeAwesome()',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    if err != nil {',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        fmt.Printf("wow: %v", err)',
                },
                {
                    kind: DiffHunkLineType.DELETED,
                    html: '        return err',
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html: '        return nil, err',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    }',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    return v.Score, nil',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '}',
                },
            ],
        },
    },
]

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/diffs/FileDiffHunks',
    decorators: [decorator],
    includeStories: ['OneDiffUnifiedHunk', 'OneDiffSplitHunk'],
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

export const OneDiffUnifiedHunk: StoryFn = args => (
    <WebStory>
        {webProps => (
            <FileDiffHunks
                diffMode="unified"
                {...webProps}
                persistLines={args.persistLines}
                fileDiffAnchor="abc"
                lineNumbers={args.lineNumbers}
                hunks={DEMO_HUNKS}
                className="abcdef"
            />
        )}
    </WebStory>
)

OneDiffUnifiedHunk.storyName = 'One diff unified hunk'

export const OneDiffSplitHunk: StoryFn = args => (
    <WebStory>
        {webProps => (
            <FileDiffHunks
                diffMode="split"
                {...webProps}
                persistLines={args.persistLines}
                fileDiffAnchor="abc"
                lineNumbers={args.lineNumbers}
                hunks={DEMO_HUNKS}
                className="abcdef"
            />
        )}
    </WebStory>
)

OneDiffSplitHunk.storyName = 'One diff split hunk'
