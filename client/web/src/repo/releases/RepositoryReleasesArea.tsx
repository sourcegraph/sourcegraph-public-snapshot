import React, { useMemo } from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { RepositoryFields } from '../../graphql-operations'
import { RepoContainerContext } from '../RepoContainer'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository tags page was not found."
    />
)

interface Props
    extends RouteComponentProps<{}>,
        Pick<RepoContainerContext, 'repo' | 'routePrefix' | 'repoHeaderContributionsLifecycleProps'>,
        BreadcrumbSetters {
    repo: RepositoryFields
    history: H.History
}

/**
 * Properties passed to all page components in the repository branches area.
 */
export interface RepositoryReleasesAreaPageProps {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

/**
 * Renders pages related to repository branches.
 */
export const RepositoryReleasesArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    useBreadcrumb,
    repo,
    routePrefix,
}) => {
    useBreadcrumb(useMemo(() => ({ key: 'tags', element: 'Tags' }), []))

    const transferProps: { repo: RepositoryFields } = {
        repo,
    }

    return (
        <div className="repository-graph-area">
            <div className="container">
                <div className="container-inner">
                    <Switch>
                        <Route
                            path={`${routePrefix}/-/tags`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            render={routeComponentProps => (
                                <RepositoryReleasesTagsPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        </div>
    )
}
