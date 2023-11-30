import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../../components/WebStory'
import { mockPreviewWorkspace } from '../../batch-spec.mock'

import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

const decorator: Decorator = story => <div className="list-group d-flex flex-column w-100">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreviewListItem',
    decorators: [decorator],
}

export default config

export const Basic: StoryFn = () => (
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

export const Cached: StoryFn = () => (
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

export const Stale: StoryFn = () => (
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

export const ReadOnly: StoryFn = () => (
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
