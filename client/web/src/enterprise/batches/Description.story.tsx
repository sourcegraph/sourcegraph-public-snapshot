import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { Description } from './Description'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/Description',
    decorators: [decorator],
}

export default config

export const Overview: Story = () => (
    <WebStory>
        {props => (
            <Description
                {...props}
                description="This is an awesome batch change. It will do great things to your codebase."
            />
        )}
    </WebStory>
)
