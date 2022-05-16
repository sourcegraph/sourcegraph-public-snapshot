import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState } from '../../../graphql-operations'

import { BatchChangeStatsCard } from './BatchChangeStatsCard'

const { add } = storiesOf('web/batches/BatchChangeStatsCard', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('draft', () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: 0,
                        closed: 0,
                        merged: 0,
                        draft: 0,
                        open: 0,
                        archived: 0,
                        total: 0,
                        unpublished: 0,
                    },
                    diffStat: { added: 0, changed: 0, deleted: 0, __typename: 'DiffStat' },
                    state: BatchChangeState.DRAFT,
                }}
            />
        )}
    </WebStory>
))

add('open', () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        closed: 10,
                        deleted: 10,
                        merged: 10,
                        draft: 5,
                        open: 10,
                        total: 100,
                        archived: 18,
                        unpublished: 55,
                    },
                    diffStat: { added: 1000, changed: 2000, deleted: 1000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
))

add('open and complete', () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: 10,
                        closed: 10,
                        merged: 80,
                        draft: 0,
                        open: 0,
                        archived: 18,
                        total: 100,
                        unpublished: 0,
                    },
                    diffStat: { added: 1000, changed: 2000, deleted: 1000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
))

add('closed', () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        closed: 10,
                        deleted: 10,
                        merged: 10,
                        draft: 0,
                        open: 10,
                        archived: 18,
                        total: 100,
                        unpublished: 60,
                    },
                    diffStat: { added: 1000, changed: 2000, deleted: 1000, __typename: 'DiffStat' },
                    state: BatchChangeState.CLOSED,
                }}
            />
        )}
    </WebStory>
))
