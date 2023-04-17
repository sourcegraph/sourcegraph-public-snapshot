import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { FileLinkProps } from './ContextFiles'
import { FIXTURE_TRANSCRIPT } from './fixtures'
import { Transcript } from './Transcript'

import styles from './Transcript.story.module.css'

const meta: ComponentMeta<typeof Transcript> = {
    title: 'cody-ui/Transcript',
    component: Transcript,

    argTypes: {
        transcript: {
            name: 'Transcript fixture',
            options: Object.keys(FIXTURE_TRANSCRIPT),
            mapping: FIXTURE_TRANSCRIPT,
            control: { type: 'select' },
        },
    },
    args: {
        transcript: FIXTURE_TRANSCRIPT[Object.keys(FIXTURE_TRANSCRIPT).sort()[0]],
    },

    decorators: [
        story => <div style={{ maxWidth: '600px', margin: '2rem auto', border: 'solid 1px #ffffff33' }}>{story()}</div>,
    ],

    parameters: {
        component: Transcript,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default meta

export const Simple: ComponentStoryObj<typeof Transcript> = {
    render: args => (
        <Transcript
            messageInProgress={{ speaker: 'assistant' }}
            transcript={args.transcript}
            fileLinkComponent={FileLink}
            transcriptItemClassName={styles.transcriptItem}
            humanTranscriptItemClassName={styles.humanTranscriptItem}
            transcriptItemParticipantClassName={styles.transcriptItemParticipant}
        />
    ),
}

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
