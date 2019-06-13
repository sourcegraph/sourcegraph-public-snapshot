import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { RepoHeaderContributionsLifecycleProps } from '../../../repo/RepoHeader'
import { RepoHeaderContributionPortal } from '../../../repo/RepoHeaderContributionPortal'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository threads page was not found."
    />
)

/**
 * Properties passed to all page components in the repository threads area.
 */
export interface RepositoryThreadsAreaContext {
    /**
     * The repository.
     */
    repo: GQL.IRepository
}

interface Props extends RepositoryThreadsAreaContext, RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {
    routePrefix: string
}

/**
 * The repository threads area.
 */
export class RepositoryThreadsArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="container mt-3">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={
                        <Link to={this.props.match.url} key="threads">
                            Threads
                        </Link>
                    }
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <Switch>
                    {/* TODO(sqs) */}
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}
