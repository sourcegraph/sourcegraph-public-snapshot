import React from 'react'
import { Link } from 'react-router-dom'

import { defaultExternalServices } from '@sourcegraph/web/src/components/externalServices/externalServices'
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
    const orgKinds = orgExternalServices?.map(service => service.kind) || []

    const { externalServices: userExternalServices } = useExternalServices(user.id)
    const userKinds = new Set((userExternalServices || []).map(service => service.kind))
    const userMissing = orgKinds.filter(kind => !userKinds.has(kind)).map(kind => defaultExternalServices[kind].title)
    return (
        <>
            {userMissing.length > 0 && (
                <Container className="mb-4">
                    <h3>Just one more step...</h3>
                    <p>
                        {
                            <>
                                Connect with{' '}
                                {userMissing.length === 1 ? userMissing[0] : userMissing[0] + ' and ' + userMissing[1]}
                            </>
                        }{' '}
                        to start searching across the {orgDisplayName} organization's private repositories on
                        Sourcegraph.
                    </p>
                    <Link className="btn btn-primary" to={`/users/${user.username}/settings/code-hosts`}>
                        Connect with {userMissing.length === 1 ? userMissing[0] : 'code hosts'}
                    </Link>
                </Container>
            )}
        </>
    )
}

export interface SearchUserNeedsCodeHost {
    orgSearchContext?: string
    user: UserContext
}

export const SearchUserNeedsCodeHost: React.FunctionComponent<SearchUserNeedsCodeHost> = ({
    orgSearchContext,
    user,
}) => {
    if (!orgSearchContext || !orgSearchContext.startsWith('@')) {
        return null
    }
    const orgName = orgSearchContext.replace('@', '')
    const org = (user.organizations?.nodes || []).find(org => org.name === orgName)
    if (!org) {
        return null
    }
    return (
        <>
            <PotentialOrgUserNeedsCodeHost user={user} org={org} />
        </>
    )
}

export interface PotentialOrgUserNeedsCodeHost {
    org: OrgData
    user: UserContext
}

export const PotentialOrgUserNeedsCodeHost: React.FunctionComponent<PotentialOrgUserNeedsCodeHost> = ({
    org,
    user,
}) => {
    const { externalServices: orgExternalServices } = useExternalServices(org.id)
    return (
        <>
            <OrgUserNeedsCodeHost
                user={user}
                orgDisplayName={org.displayName || org.name}
                orgExternalServices={orgExternalServices}
            />
        </>
    )
}
