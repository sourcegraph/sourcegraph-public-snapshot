import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'
import { MultiSelectContextProvider } from '../../MultiSelectContext'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const { add } = storiesOf('web/batches/ChangesetSelectRow', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const onSubmit = (): void => {}

add('all states', () => (
    <EnterpriseWebStory>
        {props => (
            <>
                <MultiSelectContextProvider
                    initialSelected={[]}
                    initialVisible={['id-1', 'id-2']}
                    initialTotalCount={100}
                >
                    <ChangesetSelectRow
                        {...props}
                        onSubmit={onSubmit}
                        batchChangeID="test-123"
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
                </MultiSelectContextProvider>
                <hr />
                <MultiSelectContextProvider
                    initialSelected={['id-1', 'id-2']}
                    initialVisible={['id-1', 'id-2']}
                    initialTotalCount={100}
                >
                    <ChangesetSelectRow
                        {...props}
                        onSubmit={onSubmit}
                        batchChangeID="test-123"
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
                </MultiSelectContextProvider>
                <hr />
                {/* No total count, just to make sure that's handled. */}
                <MultiSelectContextProvider initialSelected={['id-1', 'id-2']} initialVisible={['id-1', 'id-2']}>
                    <ChangesetSelectRow
                        {...props}
                        onSubmit={onSubmit}
                        batchChangeID="test-123"
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
                </MultiSelectContextProvider>
                <hr />
                <MultiSelectContextProvider
                    initialSelected="all"
                    initialVisible={['id-1', 'id-2']}
                    initialTotalCount={100}
                >
                    <ChangesetSelectRow
                        {...props}
                        onSubmit={onSubmit}
                        batchChangeID="test-123"
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
                </MultiSelectContextProvider>
                <hr />
                <MultiSelectContextProvider
                    initialSelected={['id-1', 'id-2']}
                    initialVisible={['id-1', 'id-2']}
                    initialTotalCount={100}
                >
                    <ChangesetSelectRow
                        {...props}
                        onSubmit={onSubmit}
                        batchChangeID="test-123"
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
                </MultiSelectContextProvider>
            </>
        )}
    </EnterpriseWebStory>
))
