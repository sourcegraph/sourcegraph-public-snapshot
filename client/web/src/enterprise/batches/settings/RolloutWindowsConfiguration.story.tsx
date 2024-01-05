import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { BATCH_CHANGES_SITE_CONFIGURATION } from '../backend'
import { noRolloutWindowMockResult, rolloutWindowConfigMockResult } from '../mocks'

import { RolloutWindowsConfiguration } from './RolloutWindowsConfiguration'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/RolloutWindowsConfiguration',
    decorators: [decorator],
}

export default config

export const NoRolloutWindowsConfigured: StoryFn = () => (
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

export const RolloutWindowsConfigured: StoryFn = () => (
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
