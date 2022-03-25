import React, { useMemo, useContext } from 'react'

import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import { of } from 'rxjs'

import { pluralize } from '@sourcegraph/common'
import { Button, useObservable, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { AllChangesetIDsVariables, Scalars } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { Action, DropdownButton } from '../../DropdownButton'
import { MultiSelectContext } from '../../MultiSelectContext'
import {
    queryAllChangesetIDs as _queryAllChangesetIDs,
    queryAvailableBulkOperations as _queryAvailableBulkOperations,
} from '../backend'

import { CloseChangesetsModal } from './CloseChangesetsModal'
import { CreateCommentModal } from './CreateCommentModal'
import { DetachChangesetsModal } from './DetachChangesetsModal'
import { MergeChangesetsModal } from './MergeChangesetsModal'
import { PublishChangesetsModal } from './PublishChangesetsModal'
import { ReenqueueChangesetsModal } from './ReenqueueChangesetsModal'

/**
 * Describes a possible action on the changeset list.
 */
interface ChangesetListAction extends Omit<Action, 'onTrigger'> {
    /**
     * Invoked when the action is triggered. Either onDone or onCancel need to be called
     * eventually. Can return a JSX.Element to be rendered adacent to the button (i.e. a modal).
     */
    onTrigger: (
        batchChangeID: Scalars['ID'],
        changesetIDs: Scalars['ID'][],
        onDone: () => void,
        onCancel: () => void
    ) => void | JSX.Element
}

const AVAILABLE_ACTIONS: Record<string, ChangesetListAction> = {
    DETACH: {
        type: 'detach',
        buttonLabel: 'Detach changesets',
        dropdownTitle: 'Detach changesets',
        dropdownDescription:
            "Remove the selected changesets from this batch change. Unlike archive, this can't be undone.",
        // Only show on the archived tab.
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <DetachChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
                telemetryService={eventLogger}
            />
        ),
    },
    REENQUEUE: {
        type: 'retry',
        buttonLabel: 'Retry changesets',
        dropdownTitle: 'Retry changesets',
        dropdownDescription: 'Re-enqueues the selected changesets for processing, if they failed.',
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <ReenqueueChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    COMMENT: {
        type: 'commentatore',
        buttonLabel: 'Create comment',
        dropdownTitle: 'Create comment',
        dropdownDescription:
            'Create a comment on all selected changesets. For example, you could ask people for reviews, give an update, or post a cat GIF.',
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <CreateCommentModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    MERGE: {
        type: 'merge',
        experimental: true,
        buttonLabel: 'Merge changesets',
        dropdownTitle: 'Merge changesets',
        dropdownDescription:
            'Attempt to merge all selected changesets. Some changesets may be unmergeable if there are rules preventing merge, such as CI requirements.',
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <MergeChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    CLOSE: {
        type: 'close',
        buttonLabel: 'Close changesets',
        dropdownTitle: 'Close changesets',
        dropdownDescription:
            'Attempt to close all selected changesets on the code hosts. The changesets will remain part of the batch change.',
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <CloseChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    PUBLISH: {
        type: 'publish',
        buttonLabel: 'Publish changesets',
        dropdownTitle: 'Publish changesets',
        dropdownDescription: 'Attempt to publish all selected changesets to the code hosts.',
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <PublishChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
}

export interface ChangesetSelectRowProps {
    batchChangeID: Scalars['ID']
    onSubmit: () => void
    queryArguments: Omit<AllChangesetIDsVariables, 'after'>

    /** For testing only. */
    queryAllChangesetIDs?: typeof _queryAllChangesetIDs
    queryAvailableBulkOperations?: typeof _queryAvailableBulkOperations
}

/**
 * Renders the top bar of the ChangesetList with the action buttons and the X selected
 * label. Provides select ALL functionality.
 */
export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    batchChangeID,
    onSubmit,
    queryArguments,
    queryAllChangesetIDs = _queryAllChangesetIDs,
    queryAvailableBulkOperations = _queryAvailableBulkOperations,
}) => {
    const { areAllVisibleSelected, selected, selectAll } = useContext(MultiSelectContext)

    const allChangesetIDs: string[] | undefined = useObservable(
        useMemo(() => queryAllChangesetIDs(queryArguments), [queryArguments, queryAllChangesetIDs])
    )

    /**
     * Query the backed to figure out what bulk operations can be applied
     * to the selected changesets
     */
    const availableBulkOperations = useObservable(
        useMemo(() => {
            if (Array.isArray(allChangesetIDs) && allChangesetIDs.length > 0) {
                return queryAvailableBulkOperations({ batchChange: batchChangeID, changesets: allChangesetIDs })
            }

            return of([])
        }, [queryAvailableBulkOperations, batchChangeID, allChangesetIDs])
    )

    const actions = useMemo(() => {
        if (availableBulkOperations === undefined) {
            return []
        }

        return availableBulkOperations.map(operation => {
            const action = AVAILABLE_ACTIONS[operation]
            const dropdownAction: Action = {
                ...action,
                onTrigger: (onDone, onCancel) => {
                    // Depending on the selection, the set of changeset ids to act on is different.
                    const ids = selected === 'all' ? allChangesetIDs || [] : [...selected]

                    return action.onTrigger(
                        batchChangeID,
                        ids,
                        () => {
                            onSubmit()
                            onDone()
                        },
                        onCancel
                    )
                },
            }

            return dropdownAction
        })
    }, [allChangesetIDs, availableBulkOperations, batchChangeID, onSubmit, selected])

    return (
        <>
            <div className="row align-items-center no-gutters mb-2">
                <div className="ml-2 col d-flex align-items-center">
                    <Icon className="text-muted mr-2" as={InfoCircleOutlineIcon} />
                    {selected === 'all' || allChangesetIDs?.length === selected.size ? (
                        <AllSelectedLabel count={allChangesetIDs?.length} />
                    ) : (
                        `${selected.size} ${pluralize('changeset', selected.size)}`
                    )}
                    {selected !== 'all' &&
                        areAllVisibleSelected() &&
                        allChangesetIDs &&
                        allChangesetIDs.length > selected.size && (
                            <Button className="py-0 px-1" onClick={selectAll} variant="link">
                                (Select all{allChangesetIDs !== undefined && ` ${allChangesetIDs.length}`})
                            </Button>
                        )}
                </div>
                <div className="w-100 d-block d-md-none" />
                {availableBulkOperations === undefined ? (
                    <LoadingSpinner />
                ) : (
                    <div className="m-0 col col-md-auto">
                        <div className="row no-gutters">
                            <div className="col ml-0 ml-sm-2">
                                <DropdownButton actions={actions} placeholder="Select action" />
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </>
    )
}

const AllSelectedLabel: React.FunctionComponent<{ count?: number }> = ({ count }) => {
    if (count === undefined) {
        return <>All changesets selected</>
    }

    return (
        <>
            All {count} {pluralize('changeset', count)} selected
        </>
    )
}
