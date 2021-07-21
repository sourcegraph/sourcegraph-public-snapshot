import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { RepositoryFields } from '../../../graphql-operations'
import { PatternTypeProps } from '../../../search'

import { CodeIntelRepoPage } from './CodeIntelRepoPage'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

/**
 * Properties passed to all page components in the repository code intelligence area.
 */
export interface RepositoryCodeIntelAreaPageProps
    extends ThemeProps,
        RouteComponentProps<{}>,
        BreadcrumbSetters,
        Omit<PatternTypeProps, 'setPatternType'> {
    /**
     * The active repository.
     */
    repo: RepositoryFields
    globbing: boolean
}

/**
 * Renders pages related to repository code intelligence.
 */
export const RepositoryCodeIntelArea: React.FunctionComponent<RepositoryCodeIntelAreaPageProps> = ({
    match,
    useBreadcrumb,
    ...props
}) => {
    useBreadcrumb(useMemo(() => ({ key: 'code-intelligence', element: 'Code Intelligence' }), []))

    return (
        <div className="repository-code-intelligence-area container mt-3">
            <Switch>
                <Route
                    path={`${match.url}`}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    render={routeComponentProps => <CodeIntelRepoPage {...routeComponentProps} {...props} />}
                />
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
