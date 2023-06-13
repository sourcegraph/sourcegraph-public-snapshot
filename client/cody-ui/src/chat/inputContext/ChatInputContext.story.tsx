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

export const Codebase: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about' }} />,
}

export const File: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => <ChatInputContext contextStatus={{ filePath: 'path/to/file.go' }} />,
}

export const CodebaseAndFile: ComponentStoryObj<typeof ChatInputContext> = {
    render: () => (
        <ChatInputContext contextStatus={{ codebase: 'github.com/sourcegraph/about', filePath: 'path/to/file.go' }} />
    ),
}
