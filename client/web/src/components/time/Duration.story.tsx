import { date } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'

import { H3, Code } from '@sourcegraph/wildcard'

import { WebStory } from '../WebStory'

import { Duration } from './Duration'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/Duration',
    decorators: [decorator],
}

export default config

const now = new Date()

export const Fixed: Story = () => (
    <WebStory>
        {props => (
            <Duration {...props} start={new Date(date('start', subDays(now, 1)))} end={new Date(date('end', now))} />
        )}
    </WebStory>
)

export const Active: Story = () => (
    <WebStory>
        {props => (
            <>
                <H3>Borders demonstrate how the time changing does not cause layout shift.</H3>
                <div className="d-flex">
                    <span style={{ backgroundColor: 'red', width: 100 }} />
                    <Duration {...props} start={new Date(date('start', subDays(now, 1)))} />
                    <span style={{ backgroundColor: 'red', width: 100 }} />
                </div>
                <H3 className="mt-4">
                    <Code>stableWidth=false</Code>
                </H3>
                <div className="d-flex">
                    <span style={{ backgroundColor: 'red', width: 100 }} />
                    <Duration {...props} start={new Date(date('start', subDays(now, 1)))} stableWidth={false} />
                    <span style={{ backgroundColor: 'red', width: 100 }} />
                </div>
            </>
        )}
    </WebStory>
)
