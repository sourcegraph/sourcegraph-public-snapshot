import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { HeroPage } from '../../../../components/HeroPage'
import { ThreadsAreaContext } from '../ThreadsArea'
import { ThreadsManageOverviewPage } from './ThreadsManageOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested threads management page was not found."
    />
)

interface Props extends ThreadsAreaContext, RouteComponentProps<{}> {}

/**
 * The threads management area.
 */
export class ThreadsManageArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        const context: ThreadsAreaContext = this.props
        return (
            <div className="threads-manage-area container mt-3">
                <Switch>
                    <Route
                        path={this.props.match.url}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <ThreadsManageOverviewPage {...routeComponentProps} {...context} />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
