import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'

import { ChangesetApplyPreviewNode } from './ChangesetApplyPreviewNode'
import { hiddenChangesetApplyPreviewStories } from './HiddenChangesetApplyPreviewNode.story'
import styles from './PreviewList.module.scss'
import { visibleChangesetApplyPreviewNodeStories } from './VisibleChangesetApplyPreviewNode.story'

const { add } = storiesOf('web/batches/preview/ChangesetApplyPreviewNode', module).addDecorator(story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
))

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

add('Overview', () => {
    const nodes = [
        ...Object.values(visibleChangesetApplyPreviewNodeStories(false)),
        ...Object.values(hiddenChangesetApplyPreviewStories),
    ]
    return (
        <WebStory>
            {props => (
                <>
                    {nodes.map((node, index) => (
                        <ChangesetApplyPreviewNode
                            {...props}
                            key={index}
                            node={node}
                            authenticatedUser={{
                                url: '/users/alice',
                                displayName: 'Alice',
                                username: 'alice',
                                email: 'alice@email.test',
                            }}
                            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        />
                    ))}
                </>
            )}
        </WebStory>
    )
})
