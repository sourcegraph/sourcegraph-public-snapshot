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
    orgName: string
    orgExternalServices: ListExternalServiceFields[]
    user: UserContext
}

export const OrgUserNeedsCodeHost: React.FunctionComponent<OrgUserNeedsCodeHost> = ({
    orgExternalServices,
    user,
    orgName,
}) => {
    const { externalServices: userExternalServices } = useExternalServices(user.id)
    const userKinds = new Set((userExternalServices || []).map(service => service.kind))
    const userMissing = orgExternalServices.filter(es => !userKinds.has(es.kind)).map(es => es.displayName)
 if (userMissing.length > 0) {
    const missingString = userMissing.join(' and ')
    return (
        <Container className="mb-4">
                    <h3>Just one more step...</h3>
                    <p>
                        Connect with {missingString} to start searching across the {orgName} private repositories on Sourcegraph.
                    </p>
                    <Link className="btn btn-primary" to={`/users/${user.username}/settings/code-hosts`}>
                        Connect with {missingString}
                    </Link>
                </Container>
            )
        } else {
            return null
        }
}
