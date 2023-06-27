import { ComponentStoryObj, Meta } from '@storybook/react'

import { ChatInputContext } from './ChatInputContext'

const meta: Meta = {
    title: 'cody-ui/ChatInputContext',
    component: ChatInputContext,

    decorators: [
        story => (
            <div
                style={{
                    color: 'white',
                    maxWidth: '600px',
                    margin: '2rem auto',
                    padding: '1rem',
                    border: 'solid 1px #ffffff33',
                }}
            >
                {story()}
            </div>
        ),
    ],

    parameters: {
        component: ChatInputContext,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default meta

export const Empty: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{}} />,
}

export const CodebaseIndexed: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', mode: 'embeddings', connection: true }} />,
}

export const CodebaseError: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about' }} />,
}

export const CodebaseAndFile: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => (
        <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go', mode: 'embeddings' }} />
    ),
}

export const CodebaseAndFileWithSelections: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go', mode: 'embeddings', selection: { start: { line: 0, character: 0 }, end: { line: 0, character: 0 } }}} />
            <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go', mode: 'embeddings', selection: { start: { line: 0, character: 0 }, end: { line: 1, character: 0 } }}} />
            <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go', mode: 'embeddings', selection: { start: { line: 0, character: 0 }, end: { line: 3, character: 0 } }}} />
            <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go', mode: 'embeddings', selection: { start: { line: 42, character: 333 }, end: { line: 420, character: 999 } }}} />
        </div>
    ),
}

export const File: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{ filePath: 'path/to/file.go' }} />,
}

