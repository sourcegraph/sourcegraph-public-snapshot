import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { RunServerSideModal } from './RunServerSideModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit',
    decorators: [decorator],
}

export default config

export const RunServerSideModalStory: Story = () => (
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
)

RunServerSideModalStory.storyName = 'RunServerSideModal'
