import React, { useMemo } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MinusCircleIcon from 'mdi-react/MinusCircleIcon'
import { Routes, Route } from 'react-router-dom'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { useObservable, ErrorMessage } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage, NotFoundPage } from '../../components/HeroPage'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { RouteV6Descriptor } from '../../util/contributions'

import { fetchSettingsAreaRepository } from './backend'
import { RepoSettingsSidebar, RepoSettingsSideBarGroups } from './RepoSettingsSidebar'
import { settingsGroup } from './sidebaritems'

import styles from './RepoSettingsArea.module.scss'

export interface RepoSettingsAreaRouteContext extends TelemetryProps {
    repo: SettingsAreaRepositoryFields
}

export interface RepoSettingsAreaRoute extends RouteV6Descriptor<RepoSettingsAreaRouteContext> {}

const METADATA_ROUTE: RepoSettingsAreaRoute = {
    path: '/metadata',
    render: lazyComponent(() => import('./RepoSettingsMetadataPage'), 'RepoSettingsMetadataPage'),
}

const METADATA_SIDEBAR_ITEM = {
    to: '/metadata',
    exact: true,
    label: 'Metadata',
}

interface Props extends BreadcrumbSetters, TelemetryProps {
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

    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', false)

    const memoizedRepoSettingsAreaRoutes = useMemo((): readonly RepoSettingsAreaRoute[] => {
        if (enableRepositoryMetadata) {
            return [...props.repoSettingsAreaRoutes, METADATA_ROUTE]
        }
        return props.repoSettingsAreaRoutes
    }, [enableRepositoryMetadata, props.repoSettingsAreaRoutes])

    const memoizedRepoSettingsSidebarGroups = useMemo((): RepoSettingsSideBarGroups => {
        if (!enableRepositoryMetadata) {
            return props.repoSettingsSidebarGroups
        }
        return props.repoSettingsSidebarGroups.map(group => {
            if (group.header?.label === settingsGroup?.header?.label) {
                return {
                    ...group,
                    items: [...group.items, METADATA_SIDEBAR_ITEM],
                }
            }

            return group
        })
    }, [enableRepositoryMetadata, props.repoSettingsSidebarGroups])

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
        telemetryService: props.telemetryService,
    }

    return (
        <div className={classNames('container d-flex mt-3 px-3 flex-column flex-sm-row', styles.repoSettingsArea)}>
            <RepoSettingsSidebar
                className="flex-0 mr-3"
                {...props}
                {...context}
                repoSettingsSidebarGroups={memoizedRepoSettingsSidebarGroups}
            />
            <div className="flex-bounded">
                <Routes>
                    {memoizedRepoSettingsAreaRoutes.map(
                        ({ render, path, condition = () => true }) =>
                            condition(context) && <Route key="hardcoded-key" path={path} element={render(context)} />
                    )}
                    <Route path="*" element={<NotFoundPage pageType="repository" />} />
                </Routes>
            </div>
        </div>
    )
}
