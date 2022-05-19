import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../../components/WebStory'
import { mockPreviewWorkspace } from '../../batch-spec.mock'

import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

const { add } = storiesOf(
    'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreviewListItem',
    module
).addDecorator(story => <div className="list-group d-flex flex-column w-100">{story()}</div>)

add('basic', () => (
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
))

add('cached', () => (
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
))

add('stale', () => (
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
))

add('read-only', () => (
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
))
