import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { CHANGESET_COUNTS_OVER_TIME_MOCK } from './testdata'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/BurndownChart',
    decorators: [decorator],
}

export default config

export const AllStates: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider mocks={[CHANGESET_COUNTS_OVER_TIME_MOCK]}>
                <BatchChangeBurndownChart {...webProps} batchChangeID="specid" />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllStates.storyName = 'All states'
