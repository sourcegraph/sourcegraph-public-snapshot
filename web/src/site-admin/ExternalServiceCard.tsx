import React from 'react'
import { ExternalServiceKindMetadata } from './externalServices'

interface Props extends ExternalServiceKindMetadata {}

export const ExternalServiceCard: React.FunctionComponent<Props> = (service: ExternalServiceKindMetadata) => (
    <div className="external-service-card">
        <div className={`external-service-card__icon external-service-card__icon--${service.iconBrandColor}`}>
            {service.icon}
        </div>
        <div className="external-service-card__main">
            <h2 className="external-service-card__main-header">{service.title}</h2>
            <p className="external-service-card__main-body">{service.shortDescription}</p>
        </div>
    </div>
)
