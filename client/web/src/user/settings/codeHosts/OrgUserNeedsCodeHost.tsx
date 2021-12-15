import React from 'react'
import { Link } from 'react-router-dom'

import { defaultExternalServices } from '@sourcegraph/web/src/components/externalServices/externalServices'
import { Container } from '@sourcegraph/wildcard'

import { useExternalServices } from '../../../auth/useExternalServices'
import { ListExternalServiceFields } from '../../../graphql-operations'

export interface UserContext {
    id: string
    username: string
}

export interface OrgUserNeedsCodeHost {
    orgExternalServices: ListExternalServiceFields[]
    user: UserContext
}

export const OrgUserNeedsCodeHost: React.FunctionComponent<OrgUserNeedsCodeHost> = ({ orgExternalServices, user }) => {
    const { externalServices } = useExternalServices(user.id)
    const orgKinds = orgExternalServices.map(service => service.kind)
    const userKinds = new Set((externalServices || []).map(service => service.kind))
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
                        to start searching across the AwesomeCorp organization's private repositories on Sourcegraph.
                    </p>
                    <Link className="btn btn-primary" to={`/users/${user.username}/settings/code-hosts`}>
                        Connect with {userMissing.length === 1 ? userMissing[0] : 'code hosts'}
                    </Link>
                </Container>
            )}
        </>
    )
}
