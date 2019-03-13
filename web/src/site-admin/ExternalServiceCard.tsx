import * as H from 'history'
import React from 'react'
import { LinkOrButton } from '../../../shared/src/components/LinkOrButton'
import { ExternalServiceKindMetadata } from './externalServices'

export const ExternalServiceCard: React.FunctionComponent<ExternalServiceKindMetadata> = (
    service: ExternalServiceKindMetadata
) => (
    <div className="external-service-card">
        <div className={`external-service-card__icon external-service-card__icon--${service.iconBrandColor}`}>
            {service.icon}
        </div>
        <div className="external-service-card__main">
            <h3 className="external-service-card__main-header">{service.title}</h3>
            <p className="external-service-card__main-body">{service.shortDescription}</p>
        </div>
    </div>
)

interface LinkedProps extends ExternalServiceKindMetadata {
    to: H.LocationDescriptor
}

export const LinkedExternalServiceCard: React.FunctionComponent<LinkedProps> = (props: LinkedProps) => (
    <LinkOrButton
        className={`linked-external-service-card linked-external-service-card--${props.iconBrandColor}`}
        to={props.to}
    >
        <ExternalServiceCard {...props} />
    </LinkOrButton>
)
