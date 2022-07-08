import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { GET_LICENSE_AND_USAGE_INFO } from '../list/backend'
import { getLicenseAndUsageInfoResult } from '../list/testData'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/CreateUpdateBatchChangeAlert',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export default config

const buildMocks = (isLicensed = true, hasBatchChanges = true) =>
    new WildcardMockLink([
        {
            request: { query: getDocumentNode(GET_LICENSE_AND_USAGE_INFO), variables: MATCH_ANY_PARAMETERS },
            result: { data: getLicenseAndUsageInfoResult(isLicensed, hasBatchChanges) },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

export const Create: Story = () => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={18}
                batchChange={null}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
                totalCount={1}
            />
        )}
    </WebStory>
)

export const Update: Story = () => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={199}
                batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
                totalCount={1}
            />
        )}
    </WebStory>
)

export const Disabled: Story = () => (
    <WebStory>
        {props => (
            <MultiSelectContextProvider initialSelected={['id1', 'id2']}>
                <CreateUpdateBatchChangeAlert
                    {...props}
                    specID="123"
                    toBeArchived={199}
                    batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    totalCount={1}
                />
            </MultiSelectContextProvider>
        )}
    </WebStory>
)

export const ExceedsLicense: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks(false)}>
                <CreateUpdateBatchChangeAlert
                    {...props}
                    specID="123"
                    toBeArchived={199}
                    batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    totalCount={6}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)
