import React, { useCallback, useEffect, useState } from 'react'

import { Button, Container, H3, Text } from '@sourcegraph/wildcard'

import type { OrgAreaRouteContext } from '../area/OrgArea'

import { DeleteOrgModal } from './DeleteOrgModal'

interface DeleteOrgProps extends OrgAreaRouteContext {}

/**
 * Deletes an organization.
 */
export const DeleteOrg: React.FunctionComponent<React.PropsWithChildren<DeleteOrgProps>> = props => {
    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(
        () => setShowDeleteModal(!showDeleteModal),
        [setShowDeleteModal, showDeleteModal]
    )

    useEffect(() => props.telemetryRecorder.recordEvent('org.delete', 'view'), [props.telemetryRecorder])

    return (
        <Container className="mt-3 mb-5">
            <H3 className="text-danger">Delete this organization</H3>
            <div className="d-flex justify-content-between">
                <Text className="d-flex justify-content-right">
                    This cannot be undone. Deleting an organization removes all of its resources.
                </Text>
                <Button variant="danger" size="sm" onClick={toggleDeleteModal}>
                    Delete this organization
                </Button>
                <DeleteOrgModal {...props} isOpen={showDeleteModal} toggleDeleteModal={toggleDeleteModal} />
            </div>
        </Container>
    )
}
