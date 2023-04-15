import { Meta, Story } from '@storybook/react'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { ChatMessages } from './ChatMessages'
import { FileLinkProps } from './ContextFiles'

import styles from '../../../cody-web/src/Chat.module.css'

const config: Meta = {
    title: 'cody-ui/ChatMessages',
    component: ChatMessages,

    decorators: [story => <div className="container mt-3 pb-3">{story()}</div>],

    parameters: {
        component: ChatMessages,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

const FIXTURE_TRANSCRIPT: ChatMessage[] = [
    { speaker: 'human', text: 'Hello, world!', displayText: 'Hello, world!', timestamp: '2 min ago' },
    { speaker: 'assistant', text: 'Thank you', displayText: 'Thank you', timestamp: 'now' },
]

export const Simple: Story = () => (
    <ChatMessages
        messageInProgress={null}
        transcript={FIXTURE_TRANSCRIPT}
        fileLinkComponent={FileLink}
        bubbleContentClassName={styles.bubbleContent}
        humanBubbleContentClassName={styles.humanBubbleContent}
        botBubbleContentClassName={styles.botBubbleContent}
        bubbleFooterClassName={styles.bubbleFooter}
        bubbleLoaderDotClassName={styles.bubbleLoaderDot}
    />
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
