import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { useMemo, useContext } from 'react'

import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Button } from '@sourcegraph/wildcard'

import { AllChangesetIDsVariables, Scalars } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { Action, DropdownButton } from '../../DropdownButton'
import { MultiSelectContext } from '../../MultiSelectContext'
import { queryAllChangesetIDs as _queryAllChangesetIDs } from '../backend'

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
    /* Conditionally display the action based on the given query arguments. */
    isAvailable: (queryArguments: Omit<AllChangesetIDsVariables, 'after'>) => boolean
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

const AVAILABLE_ACTIONS: ChangesetListAction[] = [
    {
        type: 'detach',
        buttonLabel: 'Detach changesets',
        dropdownTitle: 'Detach changesets',
        dropdownDescription:
            "Remove the selected changesets from this batch change. Unlike archive, this can't be undone.",
        // Only show on the archived tab.
        isAvailable: ({ onlyArchived }) => !!onlyArchived,
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
    {
        type: 'retry',
        buttonLabel: 'Retry changesets',
        dropdownTitle: 'Retry changesets',
        dropdownDescription: 'Re-enqueues the selected changesets for processing, if they failed.',
        // Only show when filtering by state === FAILED:
        isAvailable: ({ state }) => state === ChangesetState.FAILED,
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <ReenqueueChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    {
        type: 'commentatore',
        buttonLabel: 'Create comment',
        dropdownTitle: 'Create comment',
        dropdownDescription:
            'Create a comment on all selected changesets. For example, you could ask people for reviews, give an update, or post a cat GIF.',
        isAvailable: () => true,
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <CreateCommentModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    {
        type: 'merge',
        experimental: true,
        buttonLabel: 'Merge changesets',
        dropdownTitle: 'Merge changesets',
        dropdownDescription:
            'Attempt to merge all selected changesets. Some changesets may be unmergeable if there are rules preventing merge, such as CI requirements.',
        isAvailable: ({ state }) => state === ChangesetState.OPEN,
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <MergeChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    {
        type: 'close',
        buttonLabel: 'Close changesets',
        dropdownTitle: 'Close changesets',
        dropdownDescription:
            'Attempt to close all selected changesets on the code hosts. The changesets will remain part of the batch change.',
        isAvailable: ({ state }) => state === ChangesetState.OPEN || state === ChangesetState.DRAFT,
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <CloseChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
    {
        type: 'publish',
        buttonLabel: 'Publish changesets',
        dropdownTitle: 'Publish changesets',
        dropdownDescription: 'Attempt to publish all selected changesets to the code hosts.',
        isAvailable: ({ state }) => state !== ChangesetState.CLOSED,
        onTrigger: (batchChangeID, changesetIDs, onDone, onCancel) => (
            <PublishChangesetsModal
                batchChangeID={batchChangeID}
                changesetIDs={changesetIDs}
                afterCreate={onDone}
                onCancel={onCancel}
            />
        ),
    },
]

export interface ChangesetSelectRowProps {
    batchChangeID: Scalars['ID']
    onSubmit: () => void
    queryArguments: Omit<AllChangesetIDsVariables, 'after'>

    /** For testing only. */
    queryAllChangesetIDs?: typeof _queryAllChangesetIDs
    /** For testing only. */
    dropDownInitiallyOpen?: boolean
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
    dropDownInitiallyOpen = false,
}) => {
    const { areAllVisibleSelected, selected, selectAll } = useContext(MultiSelectContext)

    const allChangesetIDs: string[] | undefined = useObservable(
        useMemo(() => queryAllChangesetIDs(queryArguments), [queryArguments, queryAllChangesetIDs])
    )

    const actions = useMemo(
        () =>
            AVAILABLE_ACTIONS.filter(action => action.isAvailable(queryArguments)).map(action => {
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
            }),
        [batchChangeID, onSubmit, queryArguments, selected, allChangesetIDs]
    )

    return (
        <>
            <div className="row align-items-center no-gutters mb-2">
                <div className="ml-2 col d-flex align-items-center">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
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
                <div className="m-0 col col-md-auto">
                    <div className="row no-gutters">
                        <div className="col ml-0 ml-sm-2">
                            <DropdownButton
                                actions={actions}
                                dropdownMenuPosition="right"
                                initiallyOpen={dropDownInitiallyOpen}
                                placeholder="Select action"
                            />
                        </div>
                    </div>
                </div>
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
