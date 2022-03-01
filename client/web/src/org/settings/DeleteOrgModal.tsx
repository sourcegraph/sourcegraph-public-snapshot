import { useMutation } from '@apollo/client'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Alert, Button, Input, Link, Modal } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'

import { REMOVE_ORG_MUTATION } from './gqlQueries'

interface DeleteOrgModalProps extends OrgAreaPageProps, RouteComponentProps<{}>  {
    isOpen: boolean
    toggleDeleteModal: () => void
}

export const DeleteOrgModal: React.FunctionComponent<DeleteOrgModalProps> =  props => {
    const { org, isOpen, authenticatedUser, toggleDeleteModal } = props
    const deleteLabelId = 'deleteOrgId'
    const [organizationName, setOrgName] = useState('')
    // const history = useHistory()
    const [removeOrganization, { error }] = useMutation(REMOVE_ORG_MUTATION)
    const [errorMessage, setErrorMessage] = useState('')

    useEffect(() => { setOrgName(organizationName) }, [setOrgName, organizationName])

    const onOrgChangeName = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setErrorMessage('')
        setOrgName(event.currentTarget.value)
    }, [])

    const deleteOrg = useCallback(async () => {
        eventLogger.log('DeleteOrgClicked', org.id)

        if (org.name !== organizationName) {
            setErrorMessage('Organization name does not match')
            return
        }

        try {
            await removeOrganization({
                variables: {
                    organization: org.id,
                },
            })
        } catch(error)   {
            eventLogger.log('OrgDeletionFailed', error)
        }

    },[org, organizationName, removeOrganization])

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-org-modal"
        >
        {!authenticatedUser.viewerCanAdminister ?
            <div>
                <h3 className="text-danger" id={deleteLabelId}>
                Please contact support to delete this organization
                </h3>
                <CloseIcon
                    className="icon-inline position-absolute cursor-pointer"
                    style={{ top: '1rem', right: '1rem' }}
                    onClick={toggleDeleteModal}
                />
                <p>
                    To delete this orgnaization, please contact our support on{' '}
                    <Link target="_blank" rel="noopener noreferrer" to="mailto:support@sourcegraph.com">
                        support@sourcegraph.com
                    </Link>{' '}
                </p>
            </div> :
            <div>
                <h3 className="text-danger" id={deleteLabelId}>
                    Delete organization?
                </h3>
                <CloseIcon
                    className="icon-inline position-absolute cursor-pointer"
                    style={{ top: '1rem', right: '1rem' }}
                    onClick={toggleDeleteModal}
                />
                <p>
                    <strong>Your are going to delete { org.name } from Sourcegraph.</strong>This cannot be undone. Deleting an organization will remove all of its synced repositories from Sourcegraph, along with the organization's code insights, batch changes, code monitors and other resources.
                </p>
                <p>
                    Please type the organization's name to continue
                </p>

                {errorMessage && (
                    <Alert variant="danger">
                        { errorMessage }
                    </Alert>
                )}
                <Input
                    autoFocus={true}
                    value={organizationName}
                    onChange={onOrgChangeName}
                />
                <div className="d-flex justify-content-end mt-4">
                <Button type="button" variant="danger" onClick={deleteOrg}>
                    Delete this organization
                </Button>
            </div>
            </div>
        }
        </Modal>
    )
}
