import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { useCallback, useState } from 'react'

import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ErrorAlert } from '../../../../components/alerts'
import { CreateCommentModal } from './CreateCommentModal'

interface ChangesetListAction {
    actionType: string
    actionVerb: string
    dropdownTitle: string
    dropdownDescription: string
    isAvailable: () => boolean
    onTrigger: () => void | JSX.Element
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
        onTrigger: () => {
            return <CreateCommentModal userID={'123'} afterCreate={() => undefined} onCancel={() => undefined} />
        },
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
    selected: Set<string>
    onSubmit: () => void
    isSubmitting: boolean | Error
}

export const ChangesetSelectRow: React.FunctionComponent<ChangesetSelectRowProps> = ({
    selected,
    onSubmit,
    isSubmitting,
}) => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])
    const [selectedType, setSelectedType] = useState<string | undefined>()
    const onSelectedTypeSelect = useCallback((type: string) => {
        setSelectedType(type)
        setIsOpen(false)
    }, [])
    const [renderedElem, setRenderedElem] = useState<JSX.Element>()
    const onTriggerAction = useCallback(() => {
        const action = availableActions.find(action => action.actionType === selectedType)!
        const elem = action.onTrigger()
        if (elem !== undefined) {
            setRenderedElem(elem)
        }
        // onSubmit()
    }, [selectedType])
    const buttonLabel =
        selectedType === undefined
            ? 'Select action'
            : availableActions.find(action => action.actionType === selectedType)!.actionVerb
    return (
        <>
            {renderedElem}
            <div className="row align-items-center no-gutters">
                <div className="ml-2 col">
                    <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                    {selected.size} {pluralize('changeset', selected.size)} selected
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
                                        <>
                                            <ActionDropdownItem
                                                action={action}
                                                setSelectedType={onSelectedTypeSelect}
                                            />
                                            {index !== availableActions.length - 1 && (
                                                <div className="dropdown-divider" />
                                            )}
                                        </>
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
