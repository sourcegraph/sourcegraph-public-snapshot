import H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'

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
    shortDescription: string

    kind: GQL.ExternalServiceKind

    to?: H.LocationDescriptor
    className?: string
}

export const ExternalServiceCard: React.FunctionComponent<ExternalServiceCardProps> = ({
    title,
    icon: Icon,
    shortDescription,
    to,
    kind,
    className = '',
}) => {
    const children = (
        <div className={`p-3 d-flex align-items-start border ${className}`}>
            <Icon className="icon-inline h3 mb-0 mr-3" />
            <div className="flex-1">
                <h3 className="mb-0">{title}</h3>
                <p className="mb-0 text-muted">{shortDescription}</p>
            </div>
            {to && <ChevronRightIcon className="align-self-center" />}
        </div>
    )
    return to ? (
        <Link
            className="d-block text-left text-body text-decoration-none"
            to={to}
            data-e2e-external-service-card-link={kind}
        >
            {children}
        </Link>
    ) : (
        children
    )
}
