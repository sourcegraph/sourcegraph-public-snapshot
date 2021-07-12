import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const { add } = storiesOf('web/batches/ChangesetSelectRow', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const onSubmit = (): void => {}

add('all states', () => (
    <EnterpriseWebStory>
        {props => (
            <>
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={false}
                    batchChangeID="test-123"
                    allVisibleSelected={false}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: null,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={false}
                    batchChangeID="test-123"
                    allVisibleSelected={false}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: null,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={false}
                    batchChangeID="test-123"
                    allVisibleSelected={false}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: null,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={false}
                    batchChangeID="test-123"
                    allVisibleSelected={true}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: null,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={true}
                    batchChangeID="test-123"
                    allVisibleSelected={true}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: null,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    allSelected={false}
                    batchChangeID="test-123"
                    allVisibleSelected={false}
                    setAllSelected={() => undefined}
                    totalCount={100}
                    queryArguments={{
                        batchChange: 'test-123',
                        checkState: null,
                        onlyArchived: true,
                        onlyPublishedByThisBatchChange: null,
                        reviewState: null,
                        search: null,
                        state: null,
                    }}
                    dropDownInitiallyOpen={true}
                />
            </>
        )}
    </EnterpriseWebStory>
))
