import { Meta } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../CodeInsightsBackendStoryMock'
import { CodeInsightsGqlBackend } from '../../../core'

import { CodeInsightsGettingStartedPage } from './CodeInsightsGettingStartedPage'

const config: Meta = {
    title: 'web/insights/getting-started/CodeInsightsGettingStartedPage',
    decorators: [
        story => (
            <div className="container py-5">
                <WebStory>{() => story()}</WebStory>
            </div>
        ),
    ],
}

export default config

const codeInsightsBackend: Partial<CodeInsightsGqlBackend> = {
    // This repo doesn't actually exist, and the story shows an error `Request to
    // http://localhost:9001/.api/graphql?BulkRepositoriesSearch failed with 404 Not Found`. But it
    // still lets you see most of the page.
    getFirstExampleRepository: () => of('myrepo'),
}

export const CodeInsightsGettingStartedPageStory = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <CodeInsightsGettingStartedPage telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendStoryMock>
)
