import type { Decorator, StoryFn, Meta } from '@storybook/react'
import classNames from 'classnames'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'

import { ChangesetApplyPreviewNode } from './ChangesetApplyPreviewNode'
import { hiddenChangesetApplyPreviewStories, visibleChangesetApplyPreviewNodeStories } from './storyData'

import styles from './PreviewList.module.scss'

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

const decorator: Decorator = story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
)

const config: Meta = {
    title: 'web/batches/preview/ChangesetApplyPreviewNode',
    decorators: [decorator],
}

export default config

export const Overview: StoryFn = () => {
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
                                emails: [{ email: 'alice@email.test', isPrimary: true, verified: true }],
                            }}
                            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        />
                    ))}
                </>
            )}
        </WebStory>
    )
}
