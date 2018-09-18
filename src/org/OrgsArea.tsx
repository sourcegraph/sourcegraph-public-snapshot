import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'
import { OrgArea } from './area/OrgArea'
import { NewOrganizationPage } from './new/NewOrganizationPage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

interface Props extends RouteComponentProps<any>, ExtensionsProps {
    user: GQL.IUser | null
    isLightTheme: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
export class OrgsArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="orgs-area">
                <div className="orgs-area__content">
                    <Switch>
                        <Route path={`${this.props.match.url}/new`} component={NewOrganizationPage} exact={true} />
                        <Route path={`${this.props.match.url}/:name`} render={this.renderOrgArea} />
                        <Route component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }

    private renderOrgArea = (routeComponentProps: RouteComponentProps<any>) => (
        <OrgArea {...this.props} {...routeComponentProps} />
    )
}
