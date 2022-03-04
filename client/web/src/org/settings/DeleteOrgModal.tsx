
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { useHistory } from 'react-router'
import { RouteComponentProps } from 'react-router-dom'

import { asError } from '@sourcegraph/common'
import { Button, Input, LoadingSpinner, Modal } from '@sourcegraph/wildcard'

import { deleteOrganization } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'

interface DeleteOrgModalProps extends OrgAreaPageProps, RouteComponentProps<{}>  {
    isOpen: boolean
    toggleDeleteModal: () => void
}

export const DeleteOrgModal: React.FunctionComponent<DeleteOrgModalProps> =  props => {
    const { org, isOpen, toggleDeleteModal } = props
    const deleteLabelId = 'deleteOrgId'
    const [orgNameInput, setOrgNameInput] = useState('')
    const [orgNamesMatch, setOrgNamesMatch] = useState<boolean>()
    const [loading, setLoading] = useState<boolean | Error>(false)
    const history = useHistory()

    useEffect(() => { setOrgNameInput(orgNameInput) }, [setOrgNameInput, orgNameInput])

    const onOrgChangeName = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setOrgNameInput(event.currentTarget.value)
        setOrgNamesMatch(event.currentTarget.value === org.name)
    }, [org])

    const deleteOrg = useCallback(
        async() => {
            setLoading(true)

            try {
                await deleteOrganization(org.id, true)
                setLoading(false)
                history.push({
                    pathname: '/settings',
                })

            } catch(error)   {
                setLoading(asError(error))
                eventLogger.log('OrgDeletionFailed')
            }
        },
        [org.id, history]
    )

    if (!props.org.viewerIsMember) {
        return null
    }

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-org-modal"
        >
            <div>
                <h3 className="text-danger" id={deleteLabelId}>
                    Delete organization?
                </h3>
                <CloseIcon
                    className="icon-inline position-absolute cursor-pointer"
                    style={{ top: '1rem', right: '1rem' }}
                    onClick={toggleDeleteModal}
                />
                <p className="pt-3">
                    <strong>You are going to delete { org.name } from Sourcegraph.</strong>
                    This cannot be undone.
                    Deleting an organization will remove all of its
                    synced repositories from Sourcegraph, along with the organization's code
                    insights, batch changes, code monitors and other resources.
                </p>
                <p className="text-mutedpt-3">
                    Please type the organization's name to continue
                </p>
                <Input
                    autoFocus={true}
                    value={orgNameInput}
                    placeholder={org.name}
                    onChange={onOrgChangeName}
                    status={orgNamesMatch ? 'valid' : 'error'}                />
                <div className="d-flex justify-content-end mt-4">
                    <Button
                        type="button"
                        variant="danger"
                        onClick={deleteOrg}
                        disabled={!orgNamesMatch || loading === true}>
                        {loading === true && <LoadingSpinner />}
                        Delete this organization
                    </Button>
                </div>
            </div>
        </Modal>
    )}
