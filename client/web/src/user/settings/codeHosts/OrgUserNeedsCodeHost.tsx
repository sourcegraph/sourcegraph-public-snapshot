import React from 'react'
import { Link } from 'react-router-dom'

import { Container } from '@sourcegraph/wildcard'

import { useExternalServices } from '../../../auth/useExternalServices'
import { ListExternalServiceFields } from '../../../graphql-operations'

export interface OrgData {
    id: string
    name: string
    displayName: string | null
}

export interface UserContext {
    id: string
    username: string
    organizations?: { nodes: OrgData[] }
}

export interface OrgUserNeedsCodeHost {
    orgDisplayName: string
    orgExternalServices?: ListExternalServiceFields[]
    user: UserContext
}

export const OrgUserNeedsCodeHost: React.FunctionComponent<OrgUserNeedsCodeHost> = ({
    orgExternalServices,
    user,
    orgDisplayName,
}) => {
    const { externalServices: userExternalServices } = useExternalServices(user.id)
    const userKinds = new Set((userExternalServices || []).map(service => service.kind))
    const userMissing = (orgExternalServices || []).filter(oes => !userKinds.has(oes.kind)).map(oes => oes.displayName)
    if (userMissing.length > 0) {
        const missingString = userMissing.join(' and ')
        return (
            <Container className="mb-4">
                <h3>Just one more step...</h3>
                <p>
                    Connect with {missingString} to start searching across the {orgDisplayName} organization's private
                    repositories on Sourcegraph.
                </p>
                <Link className="btn btn-primary" to={`/users/${user.username}/settings/code-hosts`}>
                    Connect with {missingString}
                </Link>
            </Container>
        )
    }
    return null
}

export interface SearchUserNeedsCodeHost {
    orgSearchContext?: string
    user: UserContext
}

export const SearchUserNeedsCodeHost: React.FunctionComponent<SearchUserNeedsCodeHost> = ({
    orgSearchContext,
    user,
}) => {
    // If this is not an auto context (that starts with @) we show nothing
    if (!orgSearchContext || !orgSearchContext.startsWith('@')) {
        return null
    }
    const orgName = orgSearchContext.replace('@', '')
    // Is current user a member of the org that owns the context?
    const org = (user.organizations?.nodes || []).find(org => org.name === orgName)
    if (!org) {
        return null
    }
    return <PotentialOrgUserNeedsCodeHost user={user} org={org} />
}

interface PotentialOrgUserNeedsCodeHost {
    org: OrgData
    user: UserContext
}
// This is a separate component to make sure the useExternalServices hook is not used conditionally - see https://reactjs.org/docs/hooks-rules.html#only-call-hooks-at-the-top-level
const PotentialOrgUserNeedsCodeHost: React.FunctionComponent<PotentialOrgUserNeedsCodeHost> = ({ org, user }) => {
    const { externalServices: orgExternalServices } = useExternalServices(org.id)
    return (
        <OrgUserNeedsCodeHost
            user={user}
            orgDisplayName={org.displayName || org.name}
            orgExternalServices={orgExternalServices}
        />
    )
}
