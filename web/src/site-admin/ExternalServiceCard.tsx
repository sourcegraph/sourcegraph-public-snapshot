import React from 'react'
import { ExternalServiceCategory } from './externalServices'

interface Props extends ExternalServiceCategory {}
interface State {}

export class ExternalServiceCard extends React.Component<Props, State> {
    public render(): JSX.Element | null {
        const addService: ExternalServiceCategory = this.props
        return (
            <div className="external-service-card">
                <div
                    className="external-service-card__logo"
                    style={{
                        borderLeft: `2px solid ${addService.color}`,
                    }}
                >
                    {addService.icon}
                </div>
                <div className="external-service-card__main">
                    <h2 className="external-service-card__main-header">{addService.title || addService.title}</h2>
                    <p className="external-service-card__main-body">
                        {addService.shortDescription || addService.shortDescription}
                    </p>
                </div>
            </div>
        )
    }
}
