import React, { useCallback, useState } from 'react'

import { RouteComponentProps } from 'react-router-dom'

import { Button, Container, Typography } from '@sourcegraph/wildcard'

import { OrgAreaPageProps } from '../area/OrgArea'

import { DeleteOrgModal } from './DeleteOrgModal'

interface DeleteOrgProps extends OrgAreaPageProps, RouteComponentProps<{}> {}

/**
 * Deletes an organization.
 */
export const DeleteOrg: React.FunctionComponent<React.PropsWithChildren<DeleteOrgProps>> = props => {
    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(!showDeleteModal), [
        setShowDeleteModal,
        showDeleteModal,
    ])

    return (
        <Container className="mt-3 mb-5">
            <Typography.H3 className="text-danger">Delete this organization</Typography.H3>
            <div className="d-flex justify-content-between">
                <p className="d-flex justify-content-right">
                    This cannot be undone. Deleting an organization removes all of its resources.
                </p>
                <Button variant="danger" size="sm" onClick={toggleDeleteModal}>
                    Delete this organization
                </Button>
                <DeleteOrgModal {...props} isOpen={showDeleteModal} toggleDeleteModal={toggleDeleteModal} />
            </div>
        </Container>
    )
}
