import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { of } from 'rxjs'

import { BulkOperationType } from '@sourcegraph/shared/src/graphql-operations'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import {
    queryAllChangesetIDs as _queryAllChangesetIDs,
    queryAvailableBulkOperations as _queryAvailableBulkOperations,
} from '../backend'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const { add } = storiesOf('web/batches/ChangesetSelectRow', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const onSubmit = (): void => {}

const CHANGESET_IDS = new Array<string>(100).fill('fake-id').map((id, index) => `${id}-${index}`)
const HALF_CHANGESET_IDS = CHANGESET_IDS.slice(0, 50)
const queryAll100ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS)
const queryAll50ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS.slice(0, 50))

const allBulkOperations = Object.keys(BulkOperationType) as BulkOperationType[]

add('all states', () => {
    const totalChangesets = number('Total changesets', 100)
    const visibleChangesets = number('Visible changesets', 10, { range: true, min: 0, max: totalChangesets })
    const selectableChangesets = number('Selectable changesets', 100, { range: true, min: 0, max: totalChangesets })
    const selectedChangesets = number('Selected changesets', 0, { range: true, min: 0, max: selectableChangesets })

    const queryAllChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS.slice(0, selectableChangesets))
    const initialSelected = CHANGESET_IDS.slice(0, selectedChangesets)
    const initialVisible = CHANGESET_IDS.slice(0, visibleChangesets)

    const queryAvailableBulkOperations: typeof _queryAvailableBulkOperations = () => of(allBulkOperations)

    return (
        <WebStory>
            {props => (
                <>
                    <Typography.H3>Configurable</Typography.H3>
                    <MultiSelectContextProvider initialSelected={initialSelected} initialVisible={initialVisible}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAllChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">All visible, all selectable, none selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">All visible, all selectable, half selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">All visible, all selectable, all selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">All visible, half selectable, none selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">All visible, half selectable, half selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">Half visible, all selectable, none selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">Half visible, all selectable, half selected</Typography.H3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">Half visible, all selectable, all selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">Half visible, half selectable, none selected</Typography.H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
                    <Typography.H3 className="mt-3">Half visible, half selectable, half selected</Typography.H3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={queryAvailableBulkOperations}
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
