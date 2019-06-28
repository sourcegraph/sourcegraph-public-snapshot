import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../../discussions/backend'
import { useEffectAsync } from '../../../../util/useEffectAsync'
import { ChangesetIcon } from '../../../changesets/icons'
import { ChangesetCreationStatus } from '../../../changesets/preview/backend'

interface Props {
    isLoading: boolean
    onClick: (creationStatus: ChangesetCreationStatus) => void
}

const LOADING: 'loading' = 'loading'

export const CreateChangesetFromCodeActionButton: React.FunctionComponent<Props> = ({ isLoading, onClick }) => {
    const [creationStatus, setCreationStatus] = useState<ChangesetCreationStatus>(GQL.ThreadStatus.PREVIEW)
    const setCreationStatusCreate = useCallback(() => setCreationStatus(GQL.ThreadStatus.OPEN_ACTIVE), [])
    const setCreationStatusPreview = useCallback(() => setCreationStatus(GQL.ThreadStatus.PREVIEW), [])
    const onClickWithStatus = useCallback(() => onClick(creationStatus), [onClick, creationStatus])

    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffectAsync(async () => {
        try {
            setThreadsOrError(await fetchDiscussionThreads({ query: 'is:changeset is:open', first: 5 }).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [])

    return (
        <div className="btn-group" role="group">
            <button
                className="btn btn-success"
                onClick={onClickWithStatus}
                disabled={isLoading}
                style={{ minWidth: '160px' }}
            >
                <ChangesetIcon className="icon-inline mr-1" />{' '}
                {creationStatus === GQL.ThreadStatus.PREVIEW ? 'Preview' : 'Create'} changeset
            </button>
            <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
                <DropdownToggle
                    color="success"
                    className="border-left pl-1 pr-2"
                    caret={true}
                    disabled={isLoading}
                ></DropdownToggle>
                <DropdownMenu>
                    <DropdownItem onClick={setCreationStatusCreate}>
                        <div className="d-flex align-items-start">
                            <CheckIcon
                                className={`icon-inline mr-3 ${
                                    creationStatus === GQL.ThreadStatus.OPEN_ACTIVE ? '' : 'hidden'
                                }`}
                            />
                            <div>
                                <h5 className="mb-1">Create changeset</h5>
                                <span className="text-muted">
                                    Automatically creates branches and requests code reviews
                                </span>
                            </div>
                        </div>
                    </DropdownItem>
                    <DropdownItem divider={true} />
                    <DropdownItem onClick={setCreationStatusPreview}>
                        <div className="d-flex align-items-start">
                            <CheckIcon
                                className={`icon-inline mr-3 ${
                                    creationStatus === GQL.ThreadStatus.PREVIEW ? '' : 'hidden'
                                }`}
                            />
                            <div>
                                <h5 className="mb-1">Preview changeset</h5>
                                <span className="text-muted">Doesn't create a branch or request code review</span>
                            </div>
                        </div>
                    </DropdownItem>
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
                                <DropdownItem key={thread.id} onClick={() => alert('not implemented' /* TODO!(sqs) */)}>
                                    <small className="text-muted">#{thread.idWithoutKind}</small> {thread.title}
                                </DropdownItem>
                            ))}
                        </>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        </div>
    )
}
