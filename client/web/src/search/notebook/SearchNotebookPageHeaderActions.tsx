import classNames from 'classnames'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import LockIcon from 'mdi-react/LockIcon'
import StarIcon from 'mdi-react/StarIcon'
import StarOutlineIcon from 'mdi-react/StarOutlineIcon'
import WebIcon from 'mdi-react/WebIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable } from 'rxjs'
import { catchError, switchMap, tap } from 'rxjs/operators'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, useEventObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'

import {
    deleteNotebook as _deleteNotebook,
    createNotebookStar as _createNotebookStar,
    deleteNotebookStar as _deleteNotebookStar,
} from './backend'
import { DeleteNotebookModal } from './DeleteNotebookModal'
import styles from './SearchNotebookPageHeaderActions.module.scss'

export interface SearchNotebookPageHeaderActionsProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    notebookId: string
    viewerCanManage: boolean
    isPublic: boolean
    onUpdateVisibility: (isPublic: boolean) => void
    deleteNotebook: typeof _deleteNotebook
    starsCount: number
    viewerHasStarred: boolean
    createNotebookStar: typeof _createNotebookStar
    deleteNotebookStar: typeof _deleteNotebookStar
}

export const SearchNotebookPageHeaderActions: React.FunctionComponent<SearchNotebookPageHeaderActionsProps> = ({
    authenticatedUser,
    notebookId,
    viewerCanManage,
    isPublic,
    onUpdateVisibility,
    deleteNotebook,
    starsCount,
    viewerHasStarred,
    createNotebookStar,
    deleteNotebookStar,
    telemetryService,
}) => (
    <div className="d-flex align-items-center">
        <NotebookStarsButton
            disabled={authenticatedUser === null}
            notebookId={notebookId}
            starsCount={starsCount}
            viewerHasStarred={viewerHasStarred}
            createNotebookStar={createNotebookStar}
            deleteNotebookStar={deleteNotebookStar}
            telemetryService={telemetryService}
        />
        <NotebookVisibilityDropdown
            isPublic={isPublic}
            viewerCanManage={viewerCanManage}
            onUpdateVisibility={onUpdateVisibility}
            telemetryService={telemetryService}
        />
        {viewerCanManage && (
            <NotebookSettingsDropdown
                notebookId={notebookId}
                deleteNotebook={deleteNotebook}
                telemetryService={telemetryService}
            />
        )}
    </div>
)

interface NotebookSettingsDropdownProps extends TelemetryProps {
    notebookId: string
    deleteNotebook: typeof _deleteNotebook
}

const NotebookSettingsDropdown: React.FunctionComponent<NotebookSettingsDropdownProps> = ({
    notebookId,
    deleteNotebook,
    telemetryService,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(show => !show), [setShowDeleteModal])

    return (
        <>
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen} group={false}>
                <Button tag="button" outline={true} as={DropdownToggle}>
                    <DotsHorizontalIcon />
                </Button>
                <DropdownMenu right={true}>
                    <DropdownItem disabled={true}>Settings</DropdownItem>
                    <DropdownItem divider={true} />
                    <DropdownItem className="btn-danger" onClick={() => setShowDeleteModal(true)}>
                        Delete notebook
                    </DropdownItem>
                </DropdownMenu>
            </ButtonDropdown>
            <DeleteNotebookModal
                notebookId={notebookId}
                isOpen={showDeleteModal}
                toggleDeleteModal={toggleDeleteModal}
                deleteNotebook={deleteNotebook}
                telemetryService={telemetryService}
            />
        </>
    )
}

interface NotebookVisibilityDropdownProps extends TelemetryProps {
    isPublic: boolean
    viewerCanManage: boolean
    onUpdateVisibility: (isPublic: boolean) => void
}

const NotebookVisibilityDropdown: React.FunctionComponent<NotebookVisibilityDropdownProps> = ({
    isPublic: initialIsPublic,
    onUpdateVisibility,
    viewerCanManage,
    telemetryService,
}) => {
    const [isPublic, setIsPublic] = useState(initialIsPublic)
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const updateVisibility = useCallback(
        (isPublic: boolean) => {
            telemetryService.log(`SearchNotebookSet${isPublic ? 'Public' : 'Private'}Visibility`)
            onUpdateVisibility(isPublic)
            setIsPublic(isPublic)
        },
        [onUpdateVisibility, setIsPublic, telemetryService]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleOpen} group={false}>
            <Button tag="button" outline={viewerCanManage} disabled={!viewerCanManage} as={DropdownToggle}>
                {isPublic ? (
                    <span>
                        <WebIcon className="icon-inline" /> Public
                    </span>
                ) : (
                    <span>
                        <LockIcon className="icon-inline" /> Private
                    </span>
                )}
            </Button>
            <DropdownMenu right={true} className={styles.visibilityDropdownMenu}>
                <DropdownItem disabled={true}>Change notebook visibility</DropdownItem>
                <DropdownItem divider={true} />
                <DropdownItem onClick={() => updateVisibility(false)} className={styles.visibilityDropdownItem}>
                    <div>
                        <LockIcon className="icon-inline" /> Private
                    </div>
                    <div>
                        <strong>Only you</strong> will be able to view the notebook.
                    </div>
                </DropdownItem>
                <DropdownItem onClick={() => updateVisibility(true)} className={styles.visibilityDropdownItem}>
                    <div>
                        <WebIcon className="icon-inline" /> Public
                    </div>
                    <div>
                        <strong>Everyone</strong> will be able to view the notebook.
                    </div>
                </DropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}

interface NotebookStarsButtonProps extends TelemetryProps {
    notebookId: string
    disabled: boolean
    starsCount: number
    viewerHasStarred: boolean
    createNotebookStar: typeof _createNotebookStar
    deleteNotebookStar: typeof _deleteNotebookStar
}

const NotebookStarsButton: React.FunctionComponent<NotebookStarsButtonProps> = ({
    notebookId,
    disabled,
    starsCount: initialStarsCount,
    viewerHasStarred: initialViewerHasStarred,
    createNotebookStar,
    deleteNotebookStar,
    telemetryService,
}) => {
    const [starsCount, setStarsCount] = useState(initialStarsCount)
    const [viewerHasStarred, setViewerHasStarred] = useState(initialViewerHasStarred)

    const [onStarToggle] = useEventObservable(
        useCallback(
            (viewerHasStarred: Observable<boolean>) =>
                viewerHasStarred.pipe(
                    // Immediately update the UI.
                    tap(viewerHasStarred => {
                        telemetryService.log(`SearchNotebook${viewerHasStarred ? 'Remove' : 'Add'}Star`)
                        if (viewerHasStarred) {
                            setStarsCount(starsCount => starsCount - 1)
                            setViewerHasStarred(() => false)
                        } else {
                            setStarsCount(starsCount => starsCount + 1)
                            setViewerHasStarred(() => true)
                        }
                    }),
                    switchMap(viewerHasStarred =>
                        viewerHasStarred ? deleteNotebookStar(notebookId) : createNotebookStar(notebookId)
                    ),
                    catchError(() => {
                        setStarsCount(initialStarsCount)
                        setViewerHasStarred(initialViewerHasStarred)
                        return []
                    })
                ),
            [
                deleteNotebookStar,
                notebookId,
                createNotebookStar,
                initialStarsCount,
                initialViewerHasStarred,
                telemetryService,
            ]
        )
    )

    return (
        <Button
            className="d-flex align-items-center"
            outline={true}
            disabled={disabled}
            onClick={() => onStarToggle(viewerHasStarred)}
        >
            {viewerHasStarred ? (
                <StarIcon
                    className={classNames('icon-inline', styles.notebookStarIcon, styles.notebookStarIconActive)}
                />
            ) : (
                <StarOutlineIcon className={classNames('icon-inline', styles.notebookStarIcon)} />
            )}
            <span className="ml-1">{starsCount}</span>
        </Button>
    )
}
