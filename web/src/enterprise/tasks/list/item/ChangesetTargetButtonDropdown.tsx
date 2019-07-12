import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckableDropdownItem } from '../../../../components/CheckableDropdownItem'
import { fetchDiscussionThreads } from '../../../../discussions/backend'
import { useEffectAsync } from '../../../../util/useEffectAsync'
import { ChangesetIcon } from '../../../changesets/icons'
import { ChangesetCreationStatus } from '../../../changesets/preview/backend'

export interface CreateOrPreviewChangesetButtonProps {
    onClick: () => void

    disabled?: boolean
    className?: string
    buttonClassName?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A button to preview a changeset or append to an existing changeset.
 */
export const ChangesetTargetButtonDropdown: React.FunctionComponent<CreateOrPreviewChangesetButtonProps> = ({
    onClick,
    disabled,
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffectAsync(async () => {
        setThreadsOrError(LOADING)
        try {
            setThreadsOrError(await fetchDiscussionThreads({ query: 'is:changeset is:open', first: 5 }).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [])

    const [appendToExistingChangeset, setAppendToExistingChangeset] = useState<
        Pick<GQL.IDiscussionThread, 'idWithoutKind'>
    >()
    const clearAppendToExistingChangeset = useCallback(() => setAppendToExistingChangeset(undefined), [])

    return (
        <ButtonDropdown
            isOpen={isOpen}
            toggle={toggleIsOpen}
            className={`changeset-target-button-dropdown ${className}`}
        >
            <button className={`btn ${buttonClassName}`} onClick={onClick} disabled={disabled}>
                <ChangesetIcon className="icon-inline mr-1" />
                {appendToExistingChangeset === undefined
                    ? 'New changeset'
                    : `Add to changeset #${appendToExistingChangeset.idWithoutKind}`}
            </button>
            <DropdownToggle
                color="success"
                className="changeset-target-button-dropdown__dropdown-toggle pl-1 pr-2"
                caret={true}
                disabled={disabled}
            ></DropdownToggle>
            <DropdownMenu>
                <CheckableDropdownItem
                    onClick={clearAppendToExistingChangeset}
                    checked={appendToExistingChangeset === undefined}
                >
                    <h5 className="mb-1">New changeset</h5>
                    <span className="text-muted">You can preview the changes before submitting</span>
                </CheckableDropdownItem>
                <DropdownItem divider={true} />
                {threadsOrError === LOADING ? (
                    <DropdownItem header={true} className="py-1">
                        Loading changesets...
                    </DropdownItem>
                ) : isErrorLike(threadsOrError) ? (
                    <DropdownItem header={true} className="py-1">
                        Error loading changesets
                    </DropdownItem>
                ) : (
                    <>
                        <DropdownItem header={true} className="py-1">
                            Add to existing changeset...
                        </DropdownItem>
                        {threadsOrError.nodes.map(thread => (
                            // tslint:disable-next-line: jsx-no-lambda
                            <CheckableDropdownItem
                                key={thread.id}
                                onClick={() => setAppendToExistingChangeset(thread)}
                                checked={Boolean(
                                    appendToExistingChangeset && appendToExistingChangeset.id === thread.id
                                )}
                            >
                                <span className="text-muted">#{thread.idWithoutKind}</span> {thread.title}
                            </CheckableDropdownItem>
                        ))}
                    </>
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
