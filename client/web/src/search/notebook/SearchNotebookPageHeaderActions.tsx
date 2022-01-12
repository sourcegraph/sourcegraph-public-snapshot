import classNames from 'classnames'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import LockIcon from 'mdi-react/LockIcon'
import WebIcon from 'mdi-react/WebIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { deleteNotebook as _deleteNotebook } from './backend'
import { DeleteNotebookModal } from './DeleteNotebookModal'
import styles from './SearchNotebookPageHeaderActions.module.scss'

export interface SearchNotebookPageHeaderActionsProps {
    notebookId: string
    viewerCanManage: boolean
    isPublic: boolean
    onUpdateVisibility: (isPublic: boolean) => void
    deleteNotebook: typeof _deleteNotebook
}

export const SearchNotebookPageHeaderActions: React.FunctionComponent<SearchNotebookPageHeaderActionsProps> = ({
    notebookId,
    viewerCanManage,
    isPublic,
    onUpdateVisibility,
    deleteNotebook,
}) => (
    <div className="d-flex align-items-center">
        <NotebookVisibilityDropdown
            isPublic={isPublic}
            viewerCanManage={viewerCanManage}
            onUpdateVisibility={onUpdateVisibility}
        />
        {viewerCanManage && <NotebookSettingsDropdown notebookId={notebookId} deleteNotebook={deleteNotebook} />}
    </div>
)

interface NotebookSettingsDropdownProps {
    notebookId: string
    deleteNotebook: typeof _deleteNotebook
}

const NotebookSettingsDropdown: React.FunctionComponent<NotebookSettingsDropdownProps> = ({
    notebookId,
    deleteNotebook,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(show => !show), [setShowDeleteModal])

    return (
        <>
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen} group={false}>
                <DropdownToggle className="btn btn-outline" tag="button">
                    <DotsHorizontalIcon />
                </DropdownToggle>
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
            />
        </>
    )
}

interface NotebookVisibilityDropdownProps {
    isPublic: boolean
    viewerCanManage: boolean
    onUpdateVisibility: (isPublic: boolean) => void
}

const NotebookVisibilityDropdown: React.FunctionComponent<NotebookVisibilityDropdownProps> = ({
    isPublic: initialIsPublic,
    onUpdateVisibility,
    viewerCanManage,
}) => {
    const [isPublic, setIsPublic] = useState(initialIsPublic)
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(previous => !previous), [setIsOpen])

    const updateVisibility = useCallback(
        (isPublic: boolean) => {
            onUpdateVisibility(isPublic)
            setIsPublic(isPublic)
        },
        [onUpdateVisibility, setIsPublic]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleOpen} group={false}>
            <DropdownToggle
                className={classNames('btn', viewerCanManage && 'btn-outline')}
                tag="button"
                disabled={!viewerCanManage}
            >
                {isPublic ? (
                    <span>
                        <WebIcon className="icon-inline" /> Public
                    </span>
                ) : (
                    <span>
                        <LockIcon className="icon-inline" /> Private
                    </span>
                )}
            </DropdownToggle>
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
