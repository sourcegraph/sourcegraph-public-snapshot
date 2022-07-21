import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'
import { mockImportingChangesets } from '../../batch-spec.mock'
import { boolean, number, withKnobs } from '@storybook/addon-knobs'

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
                        fetchMore: () => {},
                        loading: boolean('Loading', false),
                    }}
                    isStale={boolean('Is State', false)}
                    {...props}
                />
            )}
        </WebStory>
    )
}

ImportingChangesetsPreviewListStory.storyName = 'ImportingChangesetsPreviewList'
