import React from 'react'

import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { mockWorkspace } from './WorkspacesPreview.mock'
import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

const { add } = storiesOf(
    'web/batches/CreateBatchChangePage/WorkspacesPreview/WorkspacesPreviewListItem',
    module
).addDecorator(story => <div className="list-group d-flex flex-column w-100">{story()}</div>)

add('basic', () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem {...props} isStale={false} workspace={mockWorkspace(1)} exclude={noop} />
                <WorkspacesPreviewListItem {...props} isStale={false} workspace={mockWorkspace(2)} exclude={noop} />
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
                    workspace={mockWorkspace(1, { cachedResultFound: true })}
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    workspace={mockWorkspace(2, { cachedResultFound: true })}
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
                <WorkspacesPreviewListItem {...props} isStale={true} workspace={mockWorkspace(1)} exclude={noop} />
                <WorkspacesPreviewListItem {...props} isStale={true} workspace={mockWorkspace(2)} exclude={noop} />
            </>
        )}
    </WebStory>
))
