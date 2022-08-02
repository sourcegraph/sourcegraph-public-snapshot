import { action } from '@storybook/addon-actions'
import { boolean, number, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { mockPreviewWorkspaces } from '../../batch-spec.mock'

import { WorkspacesPreviewList } from './WorkspacesPreviewList'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreviewList',
    decorators: [decorator, withKnobs],
}

export default config

export const DefaultStory: Story = () => {
    const count = number('Count', 1)
    return (
        <WebStory>
            {props => (
                <WorkspacesPreviewList
                    isStale={boolean('Stale', false)}
                    showCached={false}
                    error={undefined}
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-ignore
                    workspacesConnection={{
                        connection: {
                            totalCount: count,
                            nodes: mockPreviewWorkspaces(count),
                            pageInfo: {
                                hasNextPage: false,
                                endCursor: null,
                            },
                        },
                        hasNextPage: boolean('Has Next Page', false),
                        fetchMore: action('Fetch More'),
                    }}
                    {...props}
                />
            )}
        </WebStory>
    )
}

DefaultStory.storyName = 'default'

export const ErrorStory: Story = () => (
    <WebStory>
        {props => (
            <WorkspacesPreviewList
                showCached={false}
                error="Failed to load workspaces"
                // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                // @ts-ignore
                workspacesConnection={{
                    connection: {
                        totalCount: 0,
                        nodes: [],
                        pageInfo: {
                            hasNextPage: false,
                            endCursor: null,
                        },
                    },
                }}
                {...props}
            />
        )}
    </WebStory>
)
