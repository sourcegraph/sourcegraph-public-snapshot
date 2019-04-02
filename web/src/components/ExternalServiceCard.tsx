import H from 'history'
import React from 'react'
import { LinkOrButton } from '../../../shared/src/components/LinkOrButton'

interface ExternalServiceCardProps {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: JSX.Element | string

    /**
     * Color to display next to the icon in the external service "button"
     */
    iconBrandColor: 'github' | 'aws' | 'bitbucket' | 'gitlab' | 'gitolite' | 'phabricator' | 'git'

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription: string

    /**
     * Default display name
     */
    defaultDisplayName: string

    /**
     * Default external service configuration
     */
    defaultConfig: string
    to?: H.LocationDescriptor
}
export class ExternalServiceCard extends React.PureComponent<ExternalServiceCardProps, {}> {
    public render(): JSX.Element {
        const cardContent = (
            <div className="external-service-card">
                <div
                    className={`external-service-card__icon external-service-card__icon--${this.props.iconBrandColor}`}
                >
                    {this.props.icon}
                </div>
                <div className="external-service-card__main">
                    <h3 className="external-service-card__main-header">{this.props.title}</h3>
                    <p className="external-service-card__main-body">{this.props.shortDescription}</p>
                </div>
            </div>
        )
        if (this.props.to) {
            return (
                <LinkOrButton
                    className={`external-service-card--${this.props.iconBrandColor} linked-external-service-card--${
                        this.props.iconBrandColor
                    } linked-external-service-card`}
                    to={this.props.to}
                >
                    {cardContent}
                </LinkOrButton>
            )
        }
        return <div className={`external-service-card--${this.props.iconBrandColor}`}>{cardContent}</div>
    }
}
