import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import UserIcon from 'mdi-react/UserIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'

interface ExternalServiceCardProps {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription?: string

    kind: ExternalServiceKind

    namespace?: ExternalServiceFields['namespace']

    to?: H.LocationDescriptor
    className?: string
}

export const ExternalServiceCard: React.FunctionComponent<ExternalServiceCardProps> = ({
    title,
    icon: Icon,
    shortDescription,
    to,
    kind,
    namespace,
    className = '',
}) => {
    const children = (
        <div className={`p-3 d-flex align-items-start border ${className}`}>
            <Icon className="icon-inline h3 mb-0 mr-3" />
            <div className="flex-1">
                <h3 className={shortDescription ? 'mb-0' : 'mt-1 mb-0'}>
                    {title}
                    {namespace && (
                        <small>
                            {' '}
                            by
                            <UserIcon className="icon-inline" />
                            <Link to={namespace.url}>{namespace.namespaceName}</Link>
                        </small>
                    )}
                </h3>
                {shortDescription && <p className="mb-0 text-muted">{shortDescription}</p>}
            </div>
            {to && <ChevronRightIcon className="align-self-center" />}
        </div>
    )
    return to ? (
        <Link
            className="d-block text-left text-body text-decoration-none"
            to={to}
            data-test-external-service-card-link={kind}
        >
            {children}
        </Link>
    ) : (
        children
    )
}
