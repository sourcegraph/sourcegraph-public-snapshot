import { action } from '@storybook/addon-actions'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { mockImportingChangesets } from '../../batch-spec.mock'

import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/ImportingChangesetsPreviewList',
    decorators: [decorator],
    argTypes: {
        count: {
            name: 'Count',
            control: { type: 'number' },
        },
        isStale: {
            name: 'Stale',
            control: { type: 'boolean' },
        },
        hasNextPage: {
            name: 'Has Next Page',
            control: { type: 'boolean' },
        },
        loading: {
            name: 'Loading',
            control: { type: 'boolean' },
        },
    },
    args: {
        count: 1,
        isStale: false,
        hasNextPage: false,
        loading: false,
    },
}

export default config

export const ImportingChangesetsPreviewListStory: StoryFn = args => {
    const count = args.count
    return (
        <WebStory>
            {props => (
                <ImportingChangesetsPreviewList
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-ignore
                    importingChangesetsConnection={{
                        connection: {
                            totalCount: count,
                            error: null,
                            nodes: mockImportingChangesets(count),
                            pageInfo: {
                                hasNextPage: false,
                                endCursor: null,
                            },
                        },
                        hasNextPage: args.hasNextPage,
                        fetchMore: action('Fetch More'),
                        loading: args.loading,
                    }}
                    isStale={args.isStale}
                    {...props}
                />
            )}
        </WebStory>
    )
}

ImportingChangesetsPreviewListStory.storyName = 'ImportingChangesetsPreviewList'
