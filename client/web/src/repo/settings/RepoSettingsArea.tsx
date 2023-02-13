import React, { useMemo } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MinusCircleIcon from 'mdi-react/MinusCircleIcon'
import { Routes, Route } from 'react-router-dom-v5-compat'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable, ErrorMessage } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage, NotFoundPage } from '../../components/HeroPage'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { RouteV6Descriptor } from '../../util/contributions'

import { fetchSettingsAreaRepository } from './backend'
import { RepoSettingsSidebar, RepoSettingsSideBarGroups } from './RepoSettingsSidebar'

import styles from './RepoSettingsArea.module.scss'

export interface RepoSettingsAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: SettingsAreaRepositoryFields
}

export interface RepoSettingsAreaRoute extends RouteV6Descriptor<RepoSettingsAreaRouteContext> {}

interface Props extends BreadcrumbSetters, ThemeProps, TelemetryProps {
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: RepoSettingsSideBarGroups
    repoName: string
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
    const repoName = props.repoName
    const repoOrError = useObservable(
        useMemo(
            () => fetchSettingsAreaRepository(repoName).pipe(catchError(error => of<ErrorLike>(asError(error)))),
            [repoName]
        )
    )

    useBreadcrumb(useMemo(() => ({ key: 'settings', element: 'Settings' }), []))

    if (repoOrError === undefined) {
        return null
    }

    if (isErrorLike(repoOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoOrError.message} />} />
    }

    if (repoOrError === null) {
        return <NotFoundPage pageType="repository" />
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
        <div className={classNames('container d-flex mt-3 px-3 flex-column flex-sm-row', styles.repoSettingsArea)}>
            <RepoSettingsSidebar className="flex-0 mr-3" {...props} {...context} />
            <div className="flex-bounded">
                <Routes>
                    {props.repoSettingsAreaRoutes.map(
                        ({ render, path, condition = () => true }) =>
                            condition(context) && <Route key="hardcoded-key" path={path} element={render(context)} />
                    )}
                    <Route element={<NotFoundPage pageType="repository" />} />
                </Routes>
            </div>
        </div>
    )
}
