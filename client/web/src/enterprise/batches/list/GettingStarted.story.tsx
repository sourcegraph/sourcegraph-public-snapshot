import React from 'react'

import { Args, useMemo } from '@storybook/addons'
import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { GettingStarted } from './GettingStarted'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/GettingStarted',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
    argTypes: {
        isSourcegraphDotCom: {
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

const commonProps = (props: Args): Pick<React.ComponentProps<typeof GettingStarted>, 'isSourcegraphDotCom'> => ({
    isSourcegraphDotCom: props.isSourcegraphDotCom,
})

export const Overview: Story = args => {
    const props = { ...useMemo(() => commonProps(args), [args]) }
    return <WebStory>{() => <GettingStarted {...props} />}</WebStory>
}
