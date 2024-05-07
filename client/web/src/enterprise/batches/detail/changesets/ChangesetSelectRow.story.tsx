import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { of } from 'rxjs'

import { BulkOperationType } from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { H3 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'
import { MultiSelectContextProvider } from '../../MultiSelectContext'
import type {
    queryAllChangesetIDs as _queryAllChangesetIDs,
    queryAvailableBulkOperations as _queryAvailableBulkOperations,
} from '../backend'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const MAX_CHANGESETS = 100

const config: Meta = {
    title: 'web/batches/ChangesetSelectRow',
    decorators: [decorator],
    argTypes: {
        visibleChangesets: {
            name: 'Visible changesets',
            control: { type: 'range', min: 0, max: MAX_CHANGESETS },
        },
        selectableChangesets: {
            name: 'Selectable changesets',
            control: { type: 'range', min: 0, max: MAX_CHANGESETS },
        },
        selectedChangesets: {
            name: 'Selected changesets',
            control: { type: 'range', min: 0, max: MAX_CHANGESETS },
        },
    },
    args: {
        visibleChangesets: 10,
        selectableChangesets: 100,
        selectedChangesets: 0,
    },
}

export default config

const onSubmit = (): void => {}

const CHANGESET_IDS = new Array<string>(MAX_CHANGESETS).fill('fake-id').map((id, index) => `${id}-${index}`)
const HALF_CHANGESET_IDS = CHANGESET_IDS.slice(0, 50)
const queryAll100ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS)
const queryAll50ChangesetIDs: typeof _queryAllChangesetIDs = () => of(CHANGESET_IDS.slice(0, 50))

const allBulkOperations = Object.keys(BulkOperationType) as BulkOperationType[]

export const AllStates: StoryFn = args => {
    const queryAllChangesetIDs: typeof _queryAllChangesetIDs = () =>
        of(CHANGESET_IDS.slice(0, args.selectableChangesets))
    const initialSelected = CHANGESET_IDS.slice(0, args.selectedChangesets)
    const initialVisible = CHANGESET_IDS.slice(0, args.visibleChangesets)

    const createAvailableOperationsQuery =
        (bulkOperations: BulkOperationType[]): typeof _queryAvailableBulkOperations =>
        () =>
            of(bulkOperations)

    const allAvailableBulkOperationsQuery = createAvailableOperationsQuery(allBulkOperations)
    const commentAndDetachBulkOperationsQuery = createAvailableOperationsQuery([
        BulkOperationType.COMMENT,
        BulkOperationType.DETACH,
    ])

    return (
        <WebStory>
            {props => (
                <>
                    <H3>Configurable</H3>
                    <MultiSelectContextProvider initialSelected={initialSelected} initialVisible={initialVisible}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAllChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">All visible, all selectable, none selected</H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">All visible, all selectable, half selected</H3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">All visible, all selectable, all selected</H3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">All visible, half selectable, none selected</H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">All visible, half selectable, half selected</H3>
                    <MultiSelectContextProvider initialSelected={HALF_CHANGESET_IDS} initialVisible={CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">Half visible, all selectable, none selected</H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">Half visible, all selectable, half selected</H3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">Half visible, all selectable, all selected</H3>
                    <MultiSelectContextProvider initialSelected={CHANGESET_IDS} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll100ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">Half visible, half selectable, none selected</H3>
                    <MultiSelectContextProvider initialSelected={[]} initialVisible={HALF_CHANGESET_IDS}>
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">Half visible, half selectable, half selected</H3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={allAvailableBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                    <H3 className="mt-3">
                        Half visible, half selectable, half selected with a subset of available bulk operations
                    </H3>
                    <MultiSelectContextProvider
                        initialSelected={HALF_CHANGESET_IDS}
                        initialVisible={HALF_CHANGESET_IDS}
                    >
                        <ChangesetSelectRow
                            {...props}
                            onSubmit={onSubmit}
                            batchChangeID="test-123"
                            queryAllChangesetIDs={queryAll50ChangesetIDs}
                            queryAvailableBulkOperations={commentAndDetachBulkOperationsQuery}
                            queryArguments={{
                                batchChange: 'test-123',
                                checkState: null,
                                onlyArchived: null,
                                onlyPublishedByThisBatchChange: null,
                                reviewState: null,
                                search: null,
                                state: null,
                            }}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </MultiSelectContextProvider>
                    <hr />
                </>
            )}
        </WebStory>
    )
}

AllStates.storyName = 'All states'
