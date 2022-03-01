import React, { useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Container } from '@sourcegraph/wildcard'

import { OrgAreaPageProps } from '../area/OrgArea'

import { DeleteOrgModal } from './DeleteOrgModal'

interface DeleteOrgProps extends OrgAreaPageProps, RouteComponentProps<{}> {
    isSourcegraphDotCom: boolean,
    isOpen: boolean
}

/**
 * Delete an organization
 */
export const DeleteOrg: React.FunctionComponent<DeleteOrgProps> = props => {
    const {
        isSourcegraphDotCom,
        isOpen,
    } = props

    // const history = useHistory()

    // const onCancel = useCallback(() => {
    //     history.push('/settings')
    // }, [history])

    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(show => !show), [setShowDeleteModal])

    console.log(isSourcegraphDotCom, isOpen)
    return (
        <div className="mt-3 mb-5">
            <Container>
                <h3 className="text-danger">Delete this organization</h3>
                <div className="d-flex justify-content-between">
                    <p className="d-flex justify-content-right">This cannot be undone. Deleting an organization removes all of its resources.</p>
                    <Button
                        variant="danger"
                        size="sm"
                        onClick={toggleDeleteModal}

                    >
                        Delete this organization
                    </Button>
                    <DeleteOrgModal
                        {...props}
                        isOpen={showDeleteModal}
                        toggleDeleteModal={toggleDeleteModal}
                    />

                </div>
            </Container>
        </div>
    )
}
