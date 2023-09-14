import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { BATCH_CHANGES_SITE_CONFIGURATION } from '../backend'
import { noRolloutWindowMockResult, rolloutWindowConfigMockResult } from '../mocks'

import { RolloutWindowsConfiguration } from './RolloutWindowsConfiguration'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/RolloutWindowsConfiguration',
    decorators: [decorator],
}

export default config

export const NoRolloutWindowsConfigured: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: noRolloutWindowMockResult,
                    },
                ]}
            >
                <RolloutWindowsConfiguration />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const RolloutWindowsConfigured: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: rolloutWindowConfigMockResult,
                    },
                ]}
            >
                <RolloutWindowsConfiguration />
            </MockedTestProvider>
        )}
    </WebStory>
)
