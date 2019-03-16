import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import { HeroPage } from '../../../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../../../../repo/RepoHeader'
import { RepoHeaderContributionPortal } from '../../../../repo/RepoHeaderContributionPortal'
import { ChecksAreaContext } from '../ChecksArea'
import { ChecksManageOverviewPage } from './ChecksManageOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested checks management page was not found."
    />
)

interface Props extends ChecksAreaContext, RouteComponentProps<{}> {}

/**
 * The checks management area.
 */
export class ChecksManageArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        const context: ChecksAreaContext = this.props
        return (
            <div className="checks-manage-area">
                <Switch>
                    <Route
                        path={this.props.match.url}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <ChecksManageOverviewPage {...routeComponentProps} {...context} />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
