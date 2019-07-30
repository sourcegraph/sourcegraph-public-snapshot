import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckableDropdownItem } from '../../../../components/CheckableDropdownItem'
import { fetchDiscussionThreads } from '../../../../discussions/backend'
import { useEffectAsync } from '../../../../util/useEffectAsync'
import { ChangesetIcon } from '../../../changesetsOLD/icons'

export interface CreateOrPreviewChangesetButtonProps {
    onClick: () => void

    disabled?: boolean
    loading?: boolean
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
    loading,
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
        Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind'>
    >()
    const clearAppendToExistingChangeset = useCallback(() => setAppendToExistingChangeset(undefined), [])

    const Icon = loading ? LoadingSpinner : ChangesetIcon

    return (
        <ButtonDropdown
            isOpen={isOpen}
            toggle={toggleIsOpen}
            className={`changeset-target-button-dropdown ${className}`}
        >
            <button className={`btn ${buttonClassName}`} onClick={onClick} disabled={disabled}>
                <div
                    style={{
                        width: '22px' /* TODO!(sqs): avoid jitter bc loading spinner is not as wide as other icon */,
                    }}
                >
                    <Icon className="icon-inline mr-1" />
                </div>
                {appendToExistingChangeset === undefined
                    ? 'New changeset'
                    : `Add to changeset #${appendToExistingChangeset.idWithoutKind}`}
            </button>
            <DropdownToggle
                color="success"
                className="changeset-target-button-dropdown__dropdown-toggle pl-1 pr-2"
                caret={true}
                disabled={disabled}
            />
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
                            <CheckableDropdownItem
                                key={thread.id}
                                // tslint:disable-next-line: jsx-no-lambda
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
