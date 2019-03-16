import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { HeroPage } from '../../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../../../repo/RepoHeader'
import { CheckArea } from '../detail/CheckArea'
import { ChecksOverviewPage } from './ChecksOverviewPage'
import { ChecksManageArea } from './manage/ChecksManageArea'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested checks page was not found." />
)

/**
 * Properties passed to all page components in the checks area.
 */
export interface ChecksAreaContext {}

interface Props extends ChecksAreaContext, RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {}

/**
 * The global checks area.
 */
export class ChecksArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        const context: ChecksAreaContext = {}

        return (
            <div className="checks-area area--vertical pt-0">
                <Switch>
                    <Route
                        path={this.props.match.url}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <ChecksOverviewPage {...routeComponentProps} {...context} />}
                    />
                    <Route
                        path={`${this.props.match.url}/-/manage`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <ChecksManageArea {...routeComponentProps} {...context} />}
                    />
                    <Route
                        path={`${this.props.match.url}/:checkID`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <CheckArea {...routeComponentProps} {...context} />}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
