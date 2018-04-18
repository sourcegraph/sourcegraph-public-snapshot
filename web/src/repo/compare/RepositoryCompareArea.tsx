import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../../components/HeroPage'
import { RepoHeaderActionPortal } from '../RepoHeaderActionPortal'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepositoryCompareHeader } from './RepositoryCompareHeader'
import { RepositoryCompareOverviewPage } from './RepositoryCompareOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository comparison page was not found."
    />
)

interface Props extends RouteComponentProps<{ spec: string }> {
    repo: GQL.IRepository
}

interface State {
    error?: string
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps {
    /** The repository being compared. */
    repo: GQL.IRepository

    /** The base of the comparison. */
    base: { repoPath: string; repoID: GQLID; rev?: string | null }

    /** The head of the comparison. */
    head: { repoPath: string; repoID: GQLID; rev?: string | null }

    /** The URL route prefix for the comparison. */
    routePrefix: string
}

/**
 * Renders pages related to a repository comparison.
 */
export class RepositoryCompareArea extends React.Component<Props> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        let spec: { base: string | null; head: string | null } | null | undefined
        if (this.props.match.params.spec) {
            spec = parseComparisonSpec(this.props.match.params.spec)
        }

        const commonProps: RepositoryCompareAreaPageProps = {
            repo: this.props.repo,
            base: { repoID: this.props.repo.id, repoPath: this.props.repo.uri, rev: spec && spec.base },
            head: { repoID: this.props.repo.id, repoPath: this.props.repo.uri, rev: spec && spec.head },
            routePrefix: this.props.match.url,
        }

        return (
            <div className="repository-compare-area area--vertical">
                <RepoHeaderActionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="compare">Compare</RepoHeaderBreadcrumbNavItem>}
                />
                <RepositoryCompareHeader
                    className="area--vertical__header"
                    {...commonProps}
                    onUpdateComparisonSpec={this.onUpdateComparisonSpec}
                />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        {spec === null ? (
                            <div className="alert alert-danger">Invalid comparison specifier</div>
                        ) : (
                            <Switch>
                                <Route
                                    path={`${this.props.match.url}`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <RepositoryCompareOverviewPage {...routeComponentProps} {...commonProps} />
                                    )}
                                />
                                <Route key="hardcoded-key" component={NotFoundPage} />
                            </Switch>
                        )}
                    </div>
                </div>
            </div>
        )
    }

    private onUpdateComparisonSpec = (newBaseSpec: string, newHeadSpec: string): void => {
        this.props.history.push(
            `/${this.props.repo.uri}/-/compare${
                newBaseSpec || newHeadSpec ? `/${newBaseSpec || ''}...${newHeadSpec || ''}` : ''
            }`
        )
    }
}

function parseComparisonSpec(spec: string): { base: string | null; head: string | null } | null {
    if (!spec.includes('...')) {
        return null
    }
    const parts = spec.split('...', 2)
    return { base: parts[0] || null, head: parts[1] || null }
}
