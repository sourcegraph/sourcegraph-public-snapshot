import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'

import { GITSERVERS } from './backend'
import { SiteAdminGitserversPage } from './SiteAdminGitserversPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/Gitservers',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GITSERVERS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                gitservers: {
                    nodes: [
                        {
                            id: 'R2l0c2VydmTozNTAyIg==',
                            address: '127.0.0.1:3502',
                            freeDiskSpaceBytes: '6005174543',
                            totalDiskSpaceBytes: '994662584320',
                        },
                        {
                            id: 'R2l0c2VAuMTozNTAxIg==',
                            address: '127.0.0.1:3501',
                            freeDiskSpaceBytes: '839517454336',
                            totalDiskSpaceBytes: '994662584320',
                        },
                        {
                            id: 'R2l0c2VAuMTozNTssAxIg==',
                            address: '127.0.0.1:3503',
                            freeDiskSpaceBytes: '93059174543',
                            totalDiskSpaceBytes: '994662584320',
                        },
                    ],
                },
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const GitserversPage: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <SiteAdminGitserversPage
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

GitserversPage.storyName = 'Site Admin Gitservers Page'
