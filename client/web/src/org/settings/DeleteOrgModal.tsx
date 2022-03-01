import { useMutation } from '@apollo/client'
import React, { useCallback } from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Modal } from '@sourcegraph/wildcard'

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
               Text goes here
            </h3>

        </Modal>
    )
}
