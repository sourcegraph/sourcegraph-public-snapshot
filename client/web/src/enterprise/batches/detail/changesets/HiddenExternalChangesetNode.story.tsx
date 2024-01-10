import type { StoryFn, Meta, Decorator } from '@storybook/react'
import classNames from 'classnames'
import { addHours } from 'date-fns'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetState } from '../../../../graphql-operations'

import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'

import gridStyles from './BatchChangeChangesets.module.scss'

const decorator: Decorator = story => (
    <div className={classNames(gridStyles.batchChangeChangesetsGrid, 'p-3 container')}>{story()}</div>
)

const config: Meta = {
    title: 'web/batches/HiddenExternalChangesetNode',
    decorators: [decorator],
}

export default config

export const AllStates: StoryFn = () => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <>
                    {Object.values(ChangesetState).map((state, index) => (
                        <HiddenExternalChangesetNode
                            key={index}
                            {...props}
                            node={{
                                __typename: 'HiddenExternalChangeset',
                                id: 'somechangeset',
                                updatedAt: now.toISOString(),
                                nextSyncAt: addHours(now, 1).toISOString(),
                                state,
                            }}
                        />
                    ))}
                </>
            )}
        </WebStory>
    )
}

AllStates.storyName = 'All states'
