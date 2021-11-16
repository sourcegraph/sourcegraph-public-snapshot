import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { mockWorkspace } from './WorkspacesPreview.mock'
import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

const { add } = storiesOf(
    'web/batches/CreateBatchChangePage/WorkspacesPreview/ListItem',
    module
).addDecorator(story => <div className="list-group d-flex flex-column w-100">{story()}</div>)

add('basic', () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(1)}
                    variant="light"
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(2)}
                    variant="dark"
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
))

add('non-root path', () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(1, { path: 'path/to/workspace' })}
                    variant="light"
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(2, {
                        path: 'a/really/deeply/nested/path/that/is/super/long/and/obnoxious/like/it/just/keeps/going',
                    })}
                    variant="dark"
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
))

add('non-default branch', () => (
    <WebStory>
        {props => (
            <>
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(1, {
                        branch: {
                            id: 'not-main',
                            abbrevName: 'not-main',
                            displayName: 'not-main',
                            url: 'idk.com',
                            target: { oid: '1234' },
                        },
                    })}
                    variant="light"
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(2, {
                        branch: {
                            id: 'release 3.30',
                            abbrevName: 'release 3.30',
                            displayName: 'release 3.30',
                            url: 'idk.com',
                            target: { oid: '1234' },
                        },
                        path: '/testing/path',
                    })}
                    variant="dark"
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
                    item={mockWorkspace(1, { cachedResultFound: true })}
                    variant="light"
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={false}
                    item={mockWorkspace(2, {
                        cachedResultFound: true,
                        branch: {
                            id: 'release 3.30',
                            abbrevName: 'release 3.30',
                            displayName: 'release 3.30',
                            url: 'idk.com',
                            target: { oid: '1234' },
                        },
                        path: '/testing/path',
                    })}
                    variant="dark"
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
                    item={mockWorkspace(1)}
                    variant="light"
                    exclude={noop}
                />
                <WorkspacesPreviewListItem
                    {...props}
                    isStale={true}
                    item={mockWorkspace(2, {
                        branch: {
                            id: 'release 3.30',
                            abbrevName: 'release 3.30',
                            displayName: 'release 3.30',
                            url: 'idk.com',
                            target: { oid: '1234' },
                        },
                        path: '/testing/path',
                    })}
                    variant="dark"
                    exclude={noop}
                />
            </>
        )}
    </WebStory>
))
