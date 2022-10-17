import React, { useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { RepositoryFields } from '../../graphql-operations'

import { RepositoryStatsContributorsPage } from './RepositoryStatsContributorsPage'
import { RepositoryStatsNavbar } from './RepositoryStatsNavbar'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository stats page was not found."
    />
)

interface Props extends RouteComponentProps<{}>, BreadcrumbSetters {
    repo: RepositoryFields | undefined
    repoName: string
    globbing: boolean
}

/**
 * Properties passed to all page components in the repository stats area.
 */
export interface RepositoryStatsAreaPageProps {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

const showNavbar = false

/**
 * Renders pages related to repository stats.
 */
export const RepositoryStatsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    useBreadcrumb,
    ...props
}) => {
    useBreadcrumb(useMemo(() => ({ key: 'contributors', element: 'Contributors' }), []))

    return (
        <div className="repository-stats-area container mt-3">
            {showNavbar && <RepositoryStatsNavbar className="mb-3" repo={props.repoName} />}
            <Switch>
                <Route
                    path={`${props.match.url}/contributors`}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    render={routeComponentProps =>
                        props.repo ? (
                            <RepositoryStatsContributorsPage {...routeComponentProps} {...props} repo={props.repo} />
                        ) : (
                            <LoadingSpinner />
                        )
                    }
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
