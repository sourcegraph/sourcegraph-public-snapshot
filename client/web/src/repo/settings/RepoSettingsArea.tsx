import React, { useMemo } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import MinusCircleIcon from 'mdi-react/MinusCircleIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { RepositoryFields, SettingsAreaRepositoryFields } from '../../graphql-operations'
import { RouteDescriptor } from '../../util/contributions'

import { fetchSettingsAreaRepository } from './backend'
import { RepoSettingsSidebar, RepoSettingsSideBarGroups } from './RepoSettingsSidebar'

import styles from './RepoSettingsArea.module.scss'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository page was not found."
    />
)

export interface RepoSettingsAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: SettingsAreaRepositoryFields
}

export interface RepoSettingsAreaRoute extends RouteDescriptor<RepoSettingsAreaRouteContext> {}

interface Props extends RouteComponentProps<{}>, BreadcrumbSetters, ThemeProps, TelemetryProps {
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: RepoSettingsSideBarGroups
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * a repository's settings.
 */
export const RepoSettingsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    useBreadcrumb,

    ...props
}) => {
    const repoName = props.repo.name
    const repoOrError = useObservable(
        useMemo(() => fetchSettingsAreaRepository(repoName).pipe(catchError(error => of<ErrorLike>(asError(error)))), [
            repoName,
        ])
    )

    useBreadcrumb(useMemo(() => ({ key: 'settings', element: 'Settings' }), []))

    if (repoOrError === undefined) {
        return null
    }
    if (isErrorLike(repoOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoOrError.message} />} />
    }
    if (repoOrError === null) {
        return <NotFoundPage />
    }
    if (!repoOrError.viewerCanAdminister) {
        return (
            <HeroPage
                icon={MinusCircleIcon}
                title="Forbidden"
                subtitle="You are not authorized to view or change this repository's settings."
            />
        )
    }

    if (!props.authenticatedUser) {
        return null
    }
    const context: RepoSettingsAreaRouteContext = {
        repo: repoOrError,
        isLightTheme: props.isLightTheme,
        telemetryService: props.telemetryService,
    }

    return (
        <div className={classNames('container d-flex mt-3', styles.repoSettingsArea)}>
            <RepoSettingsSidebar className="flex-0 mr-3" {...props} {...context} />
            <div className="flex-bounded">
                <Switch>
                    {props.repoSettingsAreaRoutes.map(
                        ({ render, path, exact, condition = () => true }) =>
                            condition(context) && (
                                <Route
                                    // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    key="hardcoded-key"
                                    path={props.match.url + path}
                                    exact={exact}
                                    render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                />
                            )
                    )}
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        </div>
    )
}
