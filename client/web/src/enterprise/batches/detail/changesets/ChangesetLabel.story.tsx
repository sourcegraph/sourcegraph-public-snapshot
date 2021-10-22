import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { ChangesetLabel } from './ChangesetLabel'

const { add } = storiesOf('web/batches/ChangesetLabel', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Various labels', () => (
    <WebStory>
        {() => (
            <>
                <ChangesetLabel
                    label={{
                        text: 'Feature',
                        description: 'A feature, some descriptive text',
                        color: '93ba13',
                    }}
                />
                <ChangesetLabel
                    label={{
                        text: 'Bug',
                        description: 'A bug, some descriptive text',
                        color: 'af1302',
                    }}
                />
                <ChangesetLabel
                    label={{
                        text: 'estimate/1d',
                        description: 'An estimation, some descriptive text',
                        color: 'bfdadc',
                    }}
                />
                <ChangesetLabel
                    label={{
                        text: 'Debt',
                        description: 'Some debt, and some descriptive text',
                        color: '795548',
                    }}
                />
            </>
        )}
    </WebStory>
))
