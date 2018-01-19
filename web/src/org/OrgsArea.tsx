import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { NewOrganizationPage } from '../org/NewOrganizationPage'
import { OrgArea } from './OrgArea'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser | null
    isLightTheme: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
export class OrgsArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in.
        if (!this.props.user) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        return (
            <div className="orgs-area">
                <div className="orgs-area__content">
                    <Switch>
                        <Route path={`${this.props.match.url}/new`} component={NewOrganizationPage} exact={true} />
                        <Route path={`${this.props.match.url}/:orgName`} render={this.renderOrgArea} />
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
