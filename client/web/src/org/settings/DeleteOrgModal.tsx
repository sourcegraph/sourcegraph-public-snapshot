import React, { useCallback, useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'
import CloseIcon from 'mdi-react/CloseIcon'
import { useHistory } from 'react-router'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Input, LoadingSpinner, Modal, Icon, Typography } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'

interface DeleteOrgModalProps extends OrgAreaPageProps, RouteComponentProps<{}> {
    isOpen: boolean
    toggleDeleteModal: () => void
}

const HARD_DELETE_ORG_MUTATION = gql`
    mutation DeleteOrganization($organization: ID!, $hard: Boolean) {
        deleteOrganization(organization: $organization, hard: $hard) {
            alwaysNil
        }
    }
`

export const DeleteOrgModal: React.FunctionComponent<React.PropsWithChildren<DeleteOrgModalProps>> = props => {
    const { org, isOpen, toggleDeleteModal } = props
    const deleteLabelId = 'deleteOrgId'
    const [orgNameInput, setOrgNameInput] = useState('')
    const [orgNamesMatch, setOrgNamesMatch] = useState<boolean>()
    const history = useHistory()

    useEffect(() => {
        setOrgNameInput(orgNameInput)
    }, [setOrgNameInput, orgNameInput])

    const [deleteOrganization, { loading }] = useMutation(HARD_DELETE_ORG_MUTATION)

    const onOrgChangeName = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setOrgNameInput(event.currentTarget.value)
            setOrgNamesMatch(event.currentTarget.value === org.name)
        },
        [org]
    )

    const deleteOrg = useCallback(async () => {
        try {
            await deleteOrganization({ variables: { organization: org.id, hard: true } })
            history.push({
                pathname: '/settings',
            })
        } catch {
            eventLogger.log('OrgDeletionFailed')
        }
    }, [org, deleteOrganization, history])

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-org-modal"
        >
            <div>
                <Typography.H3 className="text-danger" id={deleteLabelId}>
                    Delete organization?
                </Typography.H3>
                <Icon
                    className="position-absolute cursor-pointer"
                    style={{ top: '1rem', right: '1rem' }}
                    onClick={toggleDeleteModal}
                    as={CloseIcon}
                />
                <p className="pt-3">
                    <strong>You are going to delete {org.name} from Sourcegraph.</strong>
                    This cannot be undone. Deleting an organization will remove all of its synced repositories from
                    Sourcegraph, along with the organization’s code insights, batch changes, code monitors and other
                    resources.
                </p>
                <Input
                    label="Please type the organization’s name to continue"
                    autoFocus={true}
                    value={orgNameInput}
                    placeholder={org.name}
                    onChange={onOrgChangeName}
                    status={orgNamesMatch ? 'valid' : 'error'}
                />
                <div className="d-flex justify-content-end mt-4">
                    <Button
                        type="button"
                        variant="danger"
                        onClick={deleteOrg}
                        disabled={!orgNamesMatch || loading === true}
                    >
                        {loading === true && <LoadingSpinner />}
                        Delete this organization
                    </Button>
                </div>
            </div>
        </Modal>
    )
}
