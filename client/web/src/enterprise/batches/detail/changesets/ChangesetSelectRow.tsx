import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { Fragment, useCallback, useState } from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { AllChangesetIDsVariables, Scalars } from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { queryAllChangesetIDs } from '../backend'

import { CreateCommentModal } from './CreateCommentModal'
import { DetachChangesetsModal } from './DetachChangesetsModal'

interface ChangesetListAction {
    actionType: string
    actionVerb: string
    dropdownTitle: string
    dropdownDescription: string
    isAvailable: (queryArguments: Omit<AllChangesetIDsVariables, 'after'>) => boolean
    onTrigger: (
        batchChangeID: Scalars['ID'],
        changesetIDs: () => Promise<Scalars['ID'][]>,
        onDone: () => void,
        onCancel: () => void
    ) => void | JSX.Element
}

const availableActions: ChangesetListAction[] = [
    {
        actionType: 'detach',
        actionVerb: 'Detach changesets',
        dropdownTitle: 'Detach changesets',
        dropdownDescription:
            "Removes the selected changesets from this batch change. Unlike archive, this can't be undone.",
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
        actionType: 'commentatore',
        actionVerb: 'Create comment',
        dropdownTitle: 'Create comment',
        dropdownDescription:
            'Create a comment on all selected changesets to ask people for reviews, or give an update.',
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
]

export interface ChangesetSelectRowProps {
    selected: Set<Scalars['ID']>
    batchChangeID: Scalars['ID']
    onSubmit: () => void
    isAllSelected: boolean
    totalCount: number
    allAllSelected: boolean
    setAllSelected: () => void
    queryArguments: Omit<AllChangesetIDsVariables, 'after'>
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    batchChangeID,
    onSubmit,
    isAllSelected,
    totalCount,
    allAllSelected,
    setAllSelected,
    queryArguments,
}) => {
    const actions = availableActions.filter(action => action.isAvailable(queryArguments))
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])
    const [selectedAction, setSelectedAction] = useState<ChangesetListAction | undefined>(() => {
        if (actions.length === 1) {
            return actions[0]
        }
        return undefined
    })
    const onSelectedTypeSelect = useCallback(
        (type: string) => {
            setSelectedAction(actions.find(action => action.actionType === type))
            setIsOpen(false)
        },
        [actions]
    )
    const [renderedElement, setRenderedElement] = useState<JSX.Element | undefined>()
    const onTriggerAction = useCallback(() => {
        if (!selectedAction) {
            return
        }
        let ids: () => Promise<Scalars['ID'][]>
        if (allAllSelected) {
            ids = () => queryAllChangesetIDs(queryArguments).toPromise()
        } else {
            ids = () => Promise.resolve([...selected])
        }
        const element = selectedAction.onTrigger(batchChangeID, ids, onSubmit, () => {
            setRenderedElement(undefined)
        })
        if (element !== undefined) {
            setRenderedElement(element)
        }
    }, [allAllSelected, batchChangeID, onSubmit, queryArguments, selected, selectedAction])

    const buttonLabel = selectedAction === undefined ? 'Select action' : selectedAction.actionVerb

    const selectedAmount = allAllSelected ? totalCount : selected.size

    return (
        <>
            {renderedElement}
            <div className="row align-items-center no-gutters">
                <div className="ml-2 col d-flex align-items-center">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                    {selectedAmount} {pluralize('changeset', selectedAmount)} selected
                    {isAllSelected && totalCount > selectedAmount && (
                        <button type="button" className="btn btn-link py-0 px-1" onClick={setAllSelected}>
                            (Select all {totalCount})
                        </button>
                    )}
                </div>
                <div className="w-100 d-block d-md-none" />
                <div className="m-0 col col-md-auto">
                    <div className="row no-gutters">
                        <div className="col my-2 ml-0 ml-sm-2">
                            <div className="btn-group">
                                <button
                                    type="button"
                                    className="btn btn-primary text-nowrap"
                                    onClick={onTriggerAction}
                                    disabled={selected.size === 0 || selectedAction === undefined}
                                >
                                    {buttonLabel}
                                </button>
                                {actions.length > 1 && (
                                    <>
                                        <button
                                            type="button"
                                            onClick={toggleIsOpen}
                                            className="btn btn-primary dropdown-toggle dropdown-toggle-split"
                                        />
                                        <div
                                            className={classNames(
                                                'dropdown-menu dropdown-menu-right',
                                                isOpen && 'show'
                                            )}
                                            style={{ minWidth: '350px' }}
                                        >
                                            {actions.map((action, index) => (
                                                <Fragment key={action.actionType}>
                                                    <ActionDropdownItem
                                                        action={action}
                                                        setSelectedType={onSelectedTypeSelect}
                                                    />
                                                    {index !== actions.length - 1 && (
                                                        <div className="dropdown-divider" />
                                                    )}
                                                </Fragment>
                                            ))}
                                        </div>
                                    </>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

const ActionDropdownItem: React.FunctionComponent<{
    setSelectedType: (type: string) => void
    action: ChangesetListAction
}> = ({ action, setSelectedType }) => {
    const onClick = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedType(action.actionType)
        },
        [setSelectedType, action.actionType]
    )
    return (
        <div className="dropdown-item">
            <button type="button" className="btn text-left" onClick={onClick}>
                <h4 className="mb-1">{action.dropdownTitle}</h4>
                <p className="text-wrap text-muted mb-0">
                    <small>{action.dropdownDescription}</small>
                </p>
            </button>
        </div>
    )
}
