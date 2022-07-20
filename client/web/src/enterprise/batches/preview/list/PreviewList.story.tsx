import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { Observable, of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { BatchSpecApplyPreviewConnectionFields, ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { GET_LICENSE_AND_USAGE_INFO } from '../../list/backend'
import { getLicenseAndUsageInfoResult } from '../../list/testData'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import { filterPublishableIDs } from '../utils'

import { PreviewList } from './PreviewList'
import { hiddenChangesetApplyPreviewStories, visibleChangesetApplyPreviewNodeStories } from './storyData'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/PreviewList',
    decorators: [decorator],
}

export default config

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

export const DefaultStory: Story = () => {
    const publicationStateSet = boolean('publication state set by spec file', false)

    const nodes: ChangesetApplyPreviewFields[] = [
        ...Object.values(visibleChangesetApplyPreviewNodeStories(publicationStateSet)),
        ...Object.values(hiddenChangesetApplyPreviewStories),
    ]

    const queryChangesetApplyPreview = (): Observable<BatchSpecApplyPreviewConnectionFields> =>
        of({
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: nodes.length,
            nodes,
        })

    const queryPublishableChangesetSpecIDs = (): Observable<string[]> =>
        of(filterPublishableIDs(Object.values(visibleChangesetApplyPreviewNodeStories(publicationStateSet))))

    return (
        <WebStory>
            {props => (
                <MultiSelectContextProvider>
                    <PreviewList
                        {...props}
                        batchSpecID="123123"
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryPublishableChangesetSpecIDs={queryPublishableChangesetSpecIDs}
                    />
                </MultiSelectContextProvider>
            )}
        </WebStory>
    )
}

DefaultStory.parameters = {
    chromatic: {
        viewports: [320, 576, 978, 1440],
    },
}

DefaultStory.storyName = 'default'

export const ExceedsLicenseStory: Story = () => {
    const publicationStateSet = boolean('publication state set by spec file', false)

    const nodes: ChangesetApplyPreviewFields[] = [
        ...Object.values(visibleChangesetApplyPreviewNodeStories(publicationStateSet)),
        ...Object.values(hiddenChangesetApplyPreviewStories),
    ]

    const queryChangesetApplyPreview = (): Observable<BatchSpecApplyPreviewConnectionFields> =>
        of({
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: nodes.length,
            nodes,
        })

    const queryPublishableChangesetSpecIDs = (): Observable<string[]> =>
        of(filterPublishableIDs(Object.values(visibleChangesetApplyPreviewNodeStories(publicationStateSet))))

    return (
        <WebStory>
            {props => (
                <MockedTestProvider
                    link={
                        new WildcardMockLink([
                            {
                                request: {
                                    query: getDocumentNode(GET_LICENSE_AND_USAGE_INFO),
                                    variables: MATCH_ANY_PARAMETERS,
                                },
                                result: { data: getLicenseAndUsageInfoResult(false, true) },
                                nMatches: Number.POSITIVE_INFINITY,
                            },
                        ])
                    }
                >
                    <MultiSelectContextProvider>
                        <PreviewList
                            {...props}
                            batchSpecID="123123"
                            authenticatedUser={{
                                url: '/users/alice',
                                displayName: 'Alice',
                                username: 'alice',
                                email: 'alice@email.test',
                            }}
                            totalCount={6}
                            queryChangesetApplyPreview={queryChangesetApplyPreview}
                            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                            queryPublishableChangesetSpecIDs={queryPublishableChangesetSpecIDs}
                        />
                    </MultiSelectContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

ExceedsLicenseStory.parameters = {
    chromatic: {
        viewports: [320, 576, 978, 1440],
    },
}

ExceedsLicenseStory.storyName = 'exceeds license'
