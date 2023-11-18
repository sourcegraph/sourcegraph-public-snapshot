import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { type Observable, of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import type { BatchSpecApplyPreviewConnectionFields, ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import { filterPublishableIDs } from '../utils'

import { PreviewList } from './PreviewList'
import { hiddenChangesetApplyPreviewStories, visibleChangesetApplyPreviewNodeStories } from './storyData'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/PreviewList',
    decorators: [decorator],
    argTypes: {
        publicationStateSet: {
            name: 'publication state set by spec file',
            control: { type: 'boolean' },
        },
    },
    args: {
        publicationStateSet: false,
    },
}

export default config

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

export const DefaultStory: StoryFn = args => {
    const publicationStateSet = args.publicationStateSet

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
                            emails: [{ email: 'alice@email.test', isPrimary: true, verified: true }],
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
