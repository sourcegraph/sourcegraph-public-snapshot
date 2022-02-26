import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { Container } from '@sourcegraph/wildcard'

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
        <div>
            <Container>And... I have a container!!</Container>
        </div>
    )}

