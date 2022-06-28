import React, { useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { RepositoryFields } from '../../../graphql-operations'

import { BatchChangeRepoPage } from './BatchChangeRepoPage'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

/**
 * Properties passed to all page components in the repository batch changes area.
 */
export interface RepositoryBatchChangesAreaPageProps extends ThemeProps, RouteComponentProps<{}>, BreadcrumbSetters {
    /**
     * The active repository.
     */
    repo: RepositoryFields
    globbing: boolean
}

/**
 * Renders pages related to repository batch changes.
 */
export const RepositoryBatchChangesArea: React.FunctionComponent<
    React.PropsWithChildren<RepositoryBatchChangesAreaPageProps>
> = ({ match, useBreadcrumb, ...props }) => {
    useBreadcrumb(useMemo(() => ({ key: 'batch-changes', element: 'Batch Changes' }), []))

    return (
        <div className="repository-batch-changes-area container mt-3">
            <Switch>
                <Route
                    path={`${match.url}`}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    render={routeComponentProps => <BatchChangeRepoPage {...routeComponentProps} {...props} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
