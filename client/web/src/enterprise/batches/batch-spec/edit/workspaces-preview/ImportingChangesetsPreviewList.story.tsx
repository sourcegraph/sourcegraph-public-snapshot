import { action } from '@storybook/addon-actions'
import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { mockImportingChangesets } from '../../batch-spec.mock'

import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/ImportingChangesetsPreviewList',
    decorators: [decorator],
    argTypes: {
        count: {
            name: 'Count',
            control: { type: 'number' },
            defaultValue: 1,
        },
        isStale: {
            name: 'Stale',
            control: { type: 'boolean' },
            defaultValue: false,
        },
        hasNextPage: {
            name: 'Has Next Page',
            control: { type: 'boolean' },
            defaultValue: false,
        },
        loading: {
            name: 'Loading',
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

export const ImportingChangesetsPreviewListStory: Story = args => {
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
