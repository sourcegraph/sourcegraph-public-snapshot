import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { Fragment, useCallback, useState } from 'react'

import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ErrorAlert } from '../../../../components/alerts'
import { AllChangesetIDsVariables, Scalars } from '../../../../graphql-operations'

import { CreateCommentModal } from './CreateCommentModal'
import { queryAllChangesetIDs } from '../backend'

interface ChangesetListAction {
    actionType: string
    actionVerb: string
    dropdownTitle: string
    dropdownDescription: string
    isAvailable: () => boolean
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
        isAvailable: () => true,
        onTrigger: () => undefined,
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
    {
        actionType: 'merge',
        actionVerb: 'Merge changesets',
        dropdownTitle: 'Merge changesets',
        dropdownDescription:
            "Merges all selected, currently open changesets on the code host. Some changesets may not be in a mergeable state, and hence won't be merged.",
        isAvailable: () => true,
        onTrigger: () => undefined,
    },
]

export interface ChangesetSelectRowProps {
    selected: Set<Scalars['ID']>
    batchChangeID: Scalars['ID']
    onSubmit: () => void
    isSubmitting: boolean | Error
    isAllSelected: boolean
    totalCount: number
    allAllSelected: boolean
    setAllSelected: () => void
    queryArgs: Omit<AllChangesetIDsVariables, 'after'>
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    batchChangeID,
    onSubmit,
    isSubmitting,
    isAllSelected,
    totalCount,
    allAllSelected,
    setAllSelected,
    queryArgs,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])
    const [selectedType, setSelectedType] = useState<string | undefined>()
    const onSelectedTypeSelect = useCallback((type: string) => {
        setSelectedType(type)
        setIsOpen(false)
    }, [])
    const [renderedElement, setRenderedElement] = useState<JSX.Element | undefined>()
    const onTriggerAction = useCallback(() => {
        const action = availableActions.find(action => action.actionType === selectedType)!
        let ids: () => Promise<Scalars['ID'][]>
        if (allAllSelected) {
            ids = () => queryAllChangesetIDs(queryArgs).toPromise()
        } else {
            ids = () => Promise.resolve([...selected])
        }
        const element = action.onTrigger(batchChangeID, ids, onSubmit, () => {
            setRenderedElement(undefined)
        })
        if (element !== undefined) {
            setRenderedElement(element)
        }
    }, [batchChangeID, onSubmit, selected, selectedType])
    const onSelectAll = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setAllSelected()
        },
        [setAllSelected]
    )
    const buttonLabel =
        selectedType === undefined
            ? 'Select action'
            : availableActions.find(action => action.actionType === selectedType)!.actionVerb

    const selectedAmount = allAllSelected ? totalCount : selected.size

    return (
        <>
            {renderedElement}
            <div className="row align-items-center no-gutters">
                <div className="ml-2 col">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                    {selectedAmount} {pluralize('changeset', selectedAmount)} selected
                    {isAllSelected && totalCount > selectedAmount && (
                        <a href="#" onClick={onSelectAll}>
                            (Select all {totalCount})
                        </a>
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
                                    disabled={
                                        selected.size === 0 || isSubmitting === true || selectedType === undefined
                                    }
                                >
                                    {buttonLabel}
                                </button>
                                <button
                                    type="button"
                                    onClick={toggleIsOpen}
                                    className="btn btn-primary dropdown-toggle dropdown-toggle-split"
                                />
                                <div
                                    className={classNames('dropdown-menu dropdown-menu-right', isOpen && 'show')}
                                    style={{ minWidth: '350px' }}
                                >
                                    {availableActions.map((action, index) => (
                                        <Fragment key={action.actionType}>
                                            <ActionDropdownItem
                                                action={action}
                                                setSelectedType={onSelectedTypeSelect}
                                            />
                                            {index !== availableActions.length - 1 && (
                                                <div className="dropdown-divider" />
                                            )}
                                        </Fragment>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {isErrorLike(isSubmitting) && <ErrorAlert error={isSubmitting} />}
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
            <a href="#" onClick={onClick}>
                <h4 className="mb-1">{action.dropdownTitle}</h4>
                <p className="text-wrap text-muted mb-0">
                    <small>{action.dropdownDescription}</small>
                </p>
            </a>
        </div>
    )
}
