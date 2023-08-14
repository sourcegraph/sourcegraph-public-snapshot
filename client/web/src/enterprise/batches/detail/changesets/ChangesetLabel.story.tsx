import type { Story, DecoratorFn, Meta } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { ChangesetLabel } from './ChangesetLabel'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/ChangesetLabel',
    decorators: [decorator],
}

export default config

export const VariousLabels: Story = () => (
    <WebStory>
        {() => (
            <>
                <ChangesetLabel
                    label={{
                        __typename: 'ChangesetLabel',
                        text: 'Feature',
                        description: 'A feature, some descriptive text',
                        color: '93ba13',
                    }}
                />
                <ChangesetLabel
                    label={{
                        __typename: 'ChangesetLabel',
                        text: 'Bug',
                        description: 'A bug, some descriptive text',
                        color: 'af1302',
                    }}
                />
                <ChangesetLabel
                    label={{
                        __typename: 'ChangesetLabel',
                        text: 'estimate/1d',
                        description: 'An estimation, some descriptive text',
                        color: 'bfdadc',
                    }}
                />
                <ChangesetLabel
                    label={{
                        __typename: 'ChangesetLabel',
                        text: 'Debt',
                        description: 'Some debt, and some descriptive text',
                        color: '795548',
                    }}
                />
            </>
        )}
    </WebStory>
)

VariousLabels.storyName = 'Various labels'
