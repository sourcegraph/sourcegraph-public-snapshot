import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerNotificationProps } from '../../../../shared/src/extensions/controller'
import { isDefined } from '../../../../shared/src/util/types'
import { BreadcrumbItem, Breadcrumbs } from '../../components/breadcrumbs/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { NamespaceAreaContext } from '../../namespaces/NamespaceArea'
import { CampaignArea } from './detail/CampaignArea'
import { CampaignsListPage } from './list/CampaignsListPage'

export interface CampaignsAreaContext extends NamespaceAreaContext {
    /** The URL to the campaigns area. */
    campaignsURL: string

    setBreadcrumbItem: (breadcrumbItem: BreadcrumbItem | undefined) => void
}

interface Props extends CampaignsAreaContext {}

/**
 * The campaigns area for a namespace's campaigns.
 */
export const CampaignsArea: React.FunctionComponent<Props> = props => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: CampaignsAreaContext = {
        ...props,
        campaignsURL: `${props.namespace.url}/campaigns`,
        setBreadcrumbItem,
    }
    const newCampaignURL = `${context.campaignsURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: props.namespace.__typename /* TODO!(sqs) */, to: props.namespace.url },
                { text: 'Campaigns', to: context.campaignsURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.campaignsURL, props.namespace.__typename, props.namespace.url]
    )

    return (
        <div className="namespace-campaigns-area">
            <Breadcrumbs items={breadcrumbItems} className="my-4" />
            <Switch>
                <Route path={props.namespace.url}>
                    <CampaignsListPage {...props} />
                </Route>
                <Route path={newCampaignURL} exact={true}>
                    <CampaignsNewPage {...context} />
                </Route>
                <Route
                    path={`${props.namespace.url}/:campaignID`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ campaignID: string }>) => (
                        <CampaignArea {...props} campaignID={routeComponentProps.match.params.campaignID} />
                    )}
                />
                <Route>
                    <HeroPage
                        icon={MapSearchIcon}
                        title="404: Not Found"
                        subtitle="Sorry, the requested page was not found."
                    />
                </Route>
            </Switch>
        </div>
    )
}
