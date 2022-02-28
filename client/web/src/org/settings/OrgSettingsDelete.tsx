import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Container } from '@sourcegraph/wildcard'

import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    className?: string
    showOrgCode: boolean
}

/**
 * Delete for org settings pages.
 */
export const OrgSettingsDelete: React.FunctionComponent<Props> = ({
    org,
}) => {
    if (!org) {
        return null
    }

    // const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember

    return (
        <div className="mt-3 mb-5">
            <Container>
                <h3 className="text-danger">Delete this organization</h3>
                <div className="d-flex justify-content-between">
                    <p className="d-flex justify-content-right">This cannot be undone. Deleting an organization removes all of its resources.</p>
                    <Button variant="danger" size="sm">Delete this organization</Button>
                </div>
            </Container>
        </div>
    )}
