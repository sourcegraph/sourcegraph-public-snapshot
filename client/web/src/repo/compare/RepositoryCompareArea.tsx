import React, { useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert, LoadingSpinner } from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { RepositoryFields, Scalars } from '../../graphql-operations'

import { RepositoryCompareHeader } from './RepositoryCompareHeader'
import { RepositoryCompareOverviewPage } from './RepositoryCompareOverviewPage'

import styles from './RepositoryCompareArea.module.scss'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository comparison page was not found."
    />
)

interface RepositoryCompareAreaProps extends RouteComponentProps<{ spec: string }>, ThemeProps, BreadcrumbSetters {
    repo?: RepositoryFields
    history: H.History
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps {
    /** The repository being compared. */
    repo: RepositoryFields

    /** The base of the comparison. */
    base: { repoName: string; repoID: Scalars['ID']; revision?: string | null }

    /** The head of the comparison. */
    head: { repoName: string; repoID: Scalars['ID']; revision?: string | null }

    /** The URL route prefix for the comparison. */
    routePrefix: string
}

/**
 * Renders pages related to a repository comparison.
 */
export const RepositoryCompareArea: React.FunctionComponent<RepositoryCompareAreaProps> = ({
    repo,
    useBreadcrumb,
    match,
    location,
    isLightTheme,
}) => {
    useBreadcrumb(useMemo(() => ({ key: 'compare', element: <>Compare</> }), []))

    let spec: { base: string | null; head: string | null } | null | undefined
    if (match.params.spec) {
        spec = parseComparisonSpec(decodeURIComponent(match.params.spec))
    }

    // Parse out the optional filePath search param, which is used to show only a single file in the compare view
    const searchParams = new URLSearchParams(location.search)
    const path = searchParams.get('filePath')

    if (!repo) {
        return <LoadingSpinner />
    }

    const commonProps: RepositoryCompareAreaPageProps = {
        repo,
        base: { repoID: repo.id, repoName: repo.name, revision: spec?.base },
        head: { repoID: repo.id, repoName: repo.name, revision: spec?.head },
        routePrefix: match.url,
    }
    return (
        <div className={classNames('container', styles.repositoryCompareArea)}>
            <RepositoryCompareHeader className="my-3" {...commonProps} />
            {spec === null ? (
                <Alert variant="danger">Invalid comparison specifier</Alert>
            ) : (
                <Switch>
                    <Route
                        path={`${match.url}`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        render={routeComponentProps => (
                            <RepositoryCompareOverviewPage
                                {...routeComponentProps}
                                {...commonProps}
                                path={path}
                                isLightTheme={isLightTheme}
                            />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            )}
        </div>
    )
}

function parseComparisonSpec(spec: string): { base: string | null; head: string | null } | null {
    if (!spec.includes('...')) {
        return null
    }
    const parts = spec.split('...', 2).map(decodeURIComponent)
    return {
        base: parts[0] || null,
        head: parts[1] || null,
    }
}
