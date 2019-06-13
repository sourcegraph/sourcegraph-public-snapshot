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
        subtitle="Sorry, the requested repository checks page was not found."
    />
)

/**
 * Properties passed to all page components in the repository checks area.
 */
export interface RepositoryChecksAreaContext {
    /**
     * The repository.
     */
    repo: GQL.IRepository
}

interface Props extends RepositoryChecksAreaContext, RouteComponentProps<{}>, RepoHeaderContributionsLifecycleProps {
    routePrefix: string
}

/**
 * The repository checks area.
 */
export class RepositoryChecksArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="container mt-3">
                <RepoHeaderContributionPortal
                    position="nav"
                    element={
                        <Link to={this.props.match.url} key="checks">
                            Checks
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
