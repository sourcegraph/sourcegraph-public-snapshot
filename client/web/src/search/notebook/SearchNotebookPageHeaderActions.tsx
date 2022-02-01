import classNames from 'classnames'
import { noop } from 'lodash'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import LockIcon from 'mdi-react/LockIcon'
import StarIcon from 'mdi-react/StarIcon'
import StarOutlineIcon from 'mdi-react/StarOutlineIcon'
import WebIcon from 'mdi-react/WebIcon'
import React, { useCallback, useState } from 'react'
import { Observable } from 'rxjs'
import { catchError, switchMap, tap } from 'rxjs/operators'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    Button,
    useEventObservable,
    MenuList,
    MenuHeader,
    Position,
} from '@sourcegraph/wildcard'

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
    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(show => !show), [setShowDeleteModal])

    return (
        <>
            <Menu>
                <MenuButton outline={true}>
                    <DotsHorizontalIcon />
                </MenuButton>
                <MenuList position={Position.bottomEnd}>
                    <MenuItem disabled={true} onSelect={noop}>
                        Settings
                    </MenuItem>
                    <MenuDivider />
                    <MenuItem className="btn-danger" onSelect={() => setShowDeleteModal(true)}>
                        Delete notebook
                    </MenuItem>
                </MenuList>
            </Menu>
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

    const updateVisibility = useCallback(
        (isPublic: boolean) => {
            telemetryService.log(`SearchNotebookSet${isPublic ? 'Public' : 'Private'}Visibility`)
            onUpdateVisibility(isPublic)
            setIsPublic(isPublic)
        },
        [onUpdateVisibility, setIsPublic, telemetryService]
    )

    return (
        <Menu>
            <MenuButton disabled={!viewerCanManage} outline={viewerCanManage}>
                {isPublic ? (
                    <span>
                        <WebIcon className="icon-inline" /> Public
                    </span>
                ) : (
                    <span>
                        <LockIcon className="icon-inline" /> Private
                    </span>
                )}
            </MenuButton>
            <MenuList position={Position.bottomEnd}>
                <MenuHeader>Change notebook visibility</MenuHeader>
                <MenuDivider />
                <MenuItem onSelect={() => updateVisibility(false)} className={styles.visibilityDropdownItem}>
                    <div>
                        <LockIcon className="icon-inline" /> Private
                    </div>
                    <div>
                        <strong>Only you</strong> will be able to view the notebook.
                    </div>
                </MenuItem>
                <MenuItem onSelect={() => updateVisibility(true)} className={styles.visibilityDropdownItem}>
                    <div>
                        <WebIcon className="icon-inline" /> Public
                    </div>
                    <div>
                        <strong>Everyone</strong> will be able to view the notebook.
                    </div>
                </MenuItem>
            </MenuList>
        </Menu>
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
