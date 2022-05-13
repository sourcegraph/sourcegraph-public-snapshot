import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { RunServerSideModal } from './RunServerSideModal'

const { add } = storiesOf('web/batches/batch-spec/edit', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('RunServerSideModal', () => (
    <WebStory>
        {props => (
            <RunServerSideModal
                setIsRunServerSideModalOpen={function (): void {
                    throw new Error('Function not implemented.')
                }}
                {...props}
            />
        )}
    </WebStory>
))
