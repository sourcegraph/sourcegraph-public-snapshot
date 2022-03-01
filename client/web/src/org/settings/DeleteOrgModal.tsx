import { useMutation } from '@apollo/client'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Modal, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'

import { REMOVE_ORG_MUTATION } from './gqlQueries'

interface DeleteOrgModalProps extends OrgAreaPageProps, RouteComponentProps<{}>  {
    isOpen: boolean
    toggleDeleteModal: () => void
}

export const DeleteOrgModal: React.FunctionComponent<DeleteOrgModalProps> =  props => {
    const { org, isOpen, toggleDeleteModal } = props
    // const LOADING = 'loading' as const
    const deleteLabelId = 'deleteOrgId'

    const [removeOrganization] = useMutation(REMOVE_ORG_MUTATION)

    const deleteOrg = useCallback(async () => {
        if (!org) {
            return
        }

        eventLogger.log('DeleteOrgClicked', org.id)
        try {
            await removeOrganization({
                variables: {
                    organization: orgId,
                },
            })
            eventLogger.log('OrgDeleted')
        } catch {
            eventLogger.log('OrgDeletionFailed')
        }
    },[org, removeOrganization])

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-org-modal"
        >
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

        </Modal>
    )
}
