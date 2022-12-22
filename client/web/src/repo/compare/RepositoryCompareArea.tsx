import * as React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert, LoadingSpinner, ErrorMessage } from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { RepositoryFields, Scalars } from '../../graphql-operations'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

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

interface RepositoryCompareAreaProps
    extends RouteComponentProps<{ spec: string }>,
        RepoHeaderContributionsLifecycleProps,
        PlatformContextProps,
        TelemetryProps,
        ThemeProps,
        SettingsCascadeProps,
        BreadcrumbSetters {
    repo?: RepositoryFields
    history: H.History
}

interface State {
    error?: string
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps extends PlatformContextProps {
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
export class RepositoryCompareArea extends React.Component<RepositoryCompareAreaProps, State> {
    private componentUpdates = new Subject<RepositoryCompareAreaProps>()

    private subscriptions = new Subscription()

    constructor(props: RepositoryCompareAreaProps) {
        super(props)
        this.state = {}
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
        this.subscriptions.add(this.props.setBreadcrumb({ key: 'compare', element: <>Compare</> }))
    }

    public shouldComponentUpdate(nextProps: Readonly<RepositoryCompareAreaProps>, nextState: Readonly<State>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return (
                <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={this.state.error} />} />
            )
        }

        let spec: { base: string | null; head: string | null } | null | undefined
        if (this.props.match.params.spec) {
            spec = parseComparisonSpec(decodeURIComponent(this.props.match.params.spec))
        }

        // Parse out the optional filePath search param, which is used to show only a single file in the compare view
        const searchParams = new URLSearchParams(this.props.location.search)
        const path = searchParams.get('filePath')

        if (!this.props.repo) {
            return <LoadingSpinner />
        }

        const commonProps: RepositoryCompareAreaPageProps = {
            repo: this.props.repo,
            base: { repoID: this.props.repo.id, repoName: this.props.repo.name, revision: spec?.base },
            head: { repoID: this.props.repo.id, repoName: this.props.repo.name, revision: spec?.head },
            routePrefix: this.props.match.url,
            platformContext: this.props.platformContext,
        }
        return (
            <div className={classNames('container', styles.repositoryCompareArea)}>
                <RepositoryCompareHeader className="my-3" {...commonProps} />
                {spec === null ? (
                    <Alert variant="danger">Invalid comparison specifier</Alert>
                ) : (
                    <Switch>
                        <Route
                            path={`${this.props.match.url}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            render={routeComponentProps => (
                                <RepositoryCompareOverviewPage
                                    {...routeComponentProps}
                                    {...commonProps}
                                    path={path}
                                    isLightTheme={this.props.isLightTheme}
                                />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                )}
            </div>
        )
    }
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
