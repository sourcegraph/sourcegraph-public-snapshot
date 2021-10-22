import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import { queryAllChangesetIDs as _queryAllChangesetIDs } from '../backend'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const { add } = storiesOf('web/batches/ChangesetSelectRow', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const onSubmit = (): void => {}

const CHANGESET_IDS = new Array<string>(100).fill('fake-id').map((id, index) => `${id}-${index}`)
const HALF_CHANGESET_IDS = CHANGESET_IDS.slice(0, 50)
const queryAll100ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS)
const queryAll50ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS.slice(0, 50))

add('all states', () => {
    const totalChangesets = number('Total changesets', 100)
    const visibleChangesets = number('Visible changesets', 10, { range: true, min: 0, max: totalChangesets })
    const selectableChangesets = number('Selectable changesets', 100, { range: true, min: 0, max: totalChangesets })
    const selectedChangesets = number('Selected changesets', 0, { range: true, min: 0, max: selectableChangesets })

    const queryAllChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS.slice(0, selectableChangesets))
    const initialSelected = CHANGESET_IDS.slice(0, selectedChangesets)
    const initialVisible = CHANGESET_IDS.slice(0, visibleChangesets)

    return (
        <WebStory>
            {props => (
                <>
                    <h3>Configurable</h3>
                    <MultiSelectContextProvider initialSelected={initialSelected} initialVisible={initialVisible}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAllChangesetIDs}
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
                    <h3 className="mt-3">All visible, all selectable, none selected</h3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">All visible, all selectable, half selected</h3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">All visible, all selectable, all selected</h3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">All visible, half selectable, none selected</h3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
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
                    <h3 className="mt-3">All visible, half selectable, half selected</h3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
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
                    <h3 className="mt-3">Half visible, all selectable, none selected</h3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">Half visible, all selectable, half selected</h3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">Half visible, all selectable, all selected</h3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
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
                    <h3 className="mt-3">Half visible, half selectable, none selected</h3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
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
                    <h3 className="mt-3">Half visible, half selectable, half selected</h3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
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
                </>
            )}
        </WebStory>
    )
})
