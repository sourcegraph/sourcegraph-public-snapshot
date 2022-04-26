import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { RunServerSideModal } from './RunServerSideModal'

const { add } = storiesOf('web/batches/create/RunServerSideModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Run Server-Side Modal', () => (
    <WebStory>
        {props => (
            <RunServerSideModal
                setIsRunServerSideModalOpen={function (condition: boolean): void {
                    throw new Error('Function not implemented.')
                }}
                name="my-batch-change"
                originalInput="name: my-batch-change"
                {...props}
            />
        )}
    </WebStory>
))
