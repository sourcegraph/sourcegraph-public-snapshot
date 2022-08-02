import { action } from '@storybook/addon-actions'
import { boolean, number, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { mockImportingChangesets } from '../../batch-spec.mock'

import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/ImportingChangesetsPreviewList',
    decorators: [decorator, withKnobs],
}

export default config

export const ImportingChangesetsPreviewListStory: Story = () => {
    const count = number('Count', 1)
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
                        hasNextPage: boolean('Has Next Page', false),
                        fetchMore: action('Fetch More'),
                        loading: boolean('Loading', false),
                    }}
                    isStale={boolean('Is Stale', false)}
                    {...props}
                />
            )}
        </WebStory>
    )
}

ImportingChangesetsPreviewListStory.storyName = 'ImportingChangesetsPreviewList'
