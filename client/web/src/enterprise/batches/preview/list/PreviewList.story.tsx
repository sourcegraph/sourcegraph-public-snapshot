import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { of, Observable } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import { BatchSpecApplyPreviewConnectionFields, ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import { filterPublishableIDs } from '../utils'

import { hiddenChangesetApplyPreviewStories } from './HiddenChangesetApplyPreviewNode.story'
import { PreviewList } from './PreviewList'
import { visibleChangesetApplyPreviewNodeStories } from './VisibleChangesetApplyPreviewNode.story'

const { add } = storiesOf('web/batches/preview', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

add('PreviewList', () => {
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
})
