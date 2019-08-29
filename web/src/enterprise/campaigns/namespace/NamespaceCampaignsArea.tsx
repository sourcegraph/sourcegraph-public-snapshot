import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isDefined } from '../../../../../shared/src/util/types'
import { BreadcrumbItem, Breadcrumbs } from '../../../components/breadcrumbs/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { ThemeProps } from '../../../theme'
import { CampaignArea } from '../detail/CampaignArea'
import { NamespaceCampaignsListPage } from './list/NamespaceCampaignsListPage'
import { CampaignsNewPage } from './new/CampaignsNewPage'

export interface NamespaceCampaignsAreaContext
    extends Pick<NamespaceAreaContext, 'namespace'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The URL to the campaigns area. */
    campaignsURL: string

    setBreadcrumbItem?: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    authenticatedUser: GQL.IUser | null
}

interface Props
    extends Pick<
            NamespaceCampaignsAreaContext,
            Exclude<keyof NamespaceCampaignsAreaContext, 'campaignsURL' | 'setBreadcrumbItem'>
        >,
        PlatformContextProps {}

/**
 * The campaigns area for a namespace.
 */
export const NamespaceCampaignsArea: React.FunctionComponent<Props> = ({ ...props }) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: NamespaceCampaignsAreaContext = {
        ...props,
        campaignsURL: `${props.namespace.url}/campaigns`,
        setBreadcrumbItem,
    }
    const newCampaignURL = `${context.campaignsURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [
                { text: props.namespace.namespaceName, to: props.namespace.url },
                { text: 'Campaigns', to: context.campaignsURL },
                breadcrumbItem,
            ].filter(isDefined),
        [breadcrumbItem, context.campaignsURL, props.namespace.namespaceName, props.namespace.url]
    )

    const breadcrumbs = <Breadcrumbs items={breadcrumbItems} className="my-4" />

    const removeHeader = (
        <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: 100%; max-width: unset !important; overflow:hidden; padding: unset; } /* TODO!(sqs): hack */`}</style>
    )

    return (
        <Switch>
            <Route path={context.campaignsURL} exact={true}>
                {breadcrumbs}
                <NamespaceCampaignsListPage {...context} newCampaignURL={newCampaignURL} />
            </Route>
            <Route
                path={newCampaignURL}
                exact={true}
                // eslint-disable-next-line react/jsx-no-bind
                render={(routeComponentProps: RouteComponentProps<{}>) => (
                    <>
                        {breadcrumbs}
                        <CampaignsNewPage {...context} {...routeComponentProps} />
                    </>
                )}
            />
            <Route
                path={`${context.campaignsURL}/:campaignID`}
                // eslint-disable-next-line react/jsx-no-bind
                render={(routeComponentProps: RouteComponentProps<{ campaignID: string }>) => (
                    <>
                        {removeHeader}
                        <CampaignArea
                            {...context}
                            {...routeComponentProps}
                            campaignID={routeComponentProps.match.params.campaignID}
                            header={breadcrumbs}
                            platformContext={props.platformContext}
                        />
                    </>
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
    )
}
