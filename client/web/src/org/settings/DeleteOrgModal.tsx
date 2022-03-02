
import { gql, useMutation } from '@apollo/client'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { useHistory } from 'react-router'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Input, Link, Modal } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'

interface DeleteOrgModalProps extends OrgAreaPageProps, RouteComponentProps<{}>  {
    isOpen: boolean
    toggleDeleteModal: () => void
}

const REMOVE_ORG_MUTATION = gql`
    mutation RemoveOrganization($organization: ID!) {
        removeOrganization(organization: $organization) {
            alwaysNil
        }
    }
`

export const DeleteOrgModal: React.FunctionComponent<DeleteOrgModalProps> =  props => {
    const { org, isOpen, authenticatedUser, toggleDeleteModal } = props
    const deleteLabelId = 'deleteOrgId'
    const [orgNameInput, setOrgNameInput] = useState('')
    const [removeOrganization] = useMutation(REMOVE_ORG_MUTATION)
    const [isOrgNameValid, setIsOrgNameValid] = useState<boolean>()
    const history = useHistory()

    useEffect(() => { setOrgNameInput(orgNameInput) }, [setOrgNameInput, orgNameInput])

    const onOrgChangeName = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setOrgNameInput(event.currentTarget.value)
        setIsOrgNameValid(event.currentTarget.value === org.name)
    }, [org])

    const deleteOrg = useCallback(
        async() => {
            try {
                await removeOrganization({
                    variables: {
                        organization: org.id,
                    },
                })
                history.push({
                    pathname: '/settings',
                })

            } catch(error)   {
                eventLogger.log('OrgDeletionFailed', error)
            }

        },
        [org, removeOrganization, history])

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-org-modal"
        >

        {!props.authenticatedUser.siteAdmin ?
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
            </div>
            :
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
                <Input
                    autoFocus={true}
                    value={orgNameInput}
                    onChange={onOrgChangeName}
                    status={isOrgNameValid === undefined ? undefined : isOrgNameValid ? 'valid' : 'error'}                />
                <div className="d-flex justify-content-end mt-4">
                <Button
                    type="button"
                    variant="danger"
                    onClick={deleteOrg}
                    disabled={!isOrgNameValid}>
                    Delete this organization
                </Button>
            </div>
            </div>
        }
        </Modal>
    )
}
