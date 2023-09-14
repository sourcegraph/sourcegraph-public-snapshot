import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../../components/WebStory'
import { mockPreviewWorkspace } from '../../batch-spec.mock'

import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

const decorator: DecoratorFn = story => <div className="list-group d-flex flex-column w-100">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreviewListItem',
    decorators: [decorator],
}

export default config

export const Basic: Story = () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    workspace={mockPreviewWorkspace(1)}
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    workspace={mockPreviewWorkspace(2)}
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
)

export const Cached: Story = () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    workspace={mockPreviewWorkspace(1, { cachedResultFound: true })}
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    workspace={mockPreviewWorkspace(2, { cachedResultFound: true })}
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
)

export const Stale: Story = () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={true}
                    workspace={mockPreviewWorkspace(1)}
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={true}
                    workspace={mockPreviewWorkspace(2)}
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
)

export const ReadOnly: Story = () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    isReadOnly={true}
                    workspace={mockPreviewWorkspace(1)}
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    isReadOnly={true}
                    workspace={mockPreviewWorkspace(2)}
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
)

ReadOnly.storyName = 'read-only'
