import * as H from 'history'
import { escapeRegExp, isEqual } from 'lodash'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { SymbolIcon } from '../../../shared/src/symbols/SymbolIcon'
import { FilteredConnection } from '../components/FilteredConnection'
import { fetchSymbols } from '../symbols/backend'
import { parseBrowserRepoURL } from '../util/url'

function symbolIsActive(symbolLocation: string, currentLocation: H.Location): boolean {
    const current = parseBrowserRepoURL(H.createPath(currentLocation))
    const symbol = parseBrowserRepoURL(symbolLocation)
    return (
        current.repoName === symbol.repoName &&
        current.rev === symbol.rev &&
        current.filePath === symbol.filePath &&
        isEqual(current.position, symbol.position)
    )
}

const symbolIsActiveTrue = (): boolean => true
const symbolIsActiveFalse = (): boolean => false

interface SymbolNodeProps {
    node: GQL.ISymbol
    location: H.Location
}

const SymbolNode: React.FunctionComponent<SymbolNodeProps> = ({ node, location }) => {
    const isActiveFunc = symbolIsActive(node.url, location) ? symbolIsActiveTrue : symbolIsActiveFalse
    return (
        <li className="repo-rev-sidebar-symbols-node">
            <NavLink
                to={node.url}
                isActive={isActiveFunc}
                className="repo-rev-sidebar-symbols-node__link e2e-symbol-link"
                activeClassName="repo-rev-sidebar-symbols-node__link--active"
            >
                <SymbolIcon kind={node.kind} className="icon-inline mr-1 e2e-symbol-icon" />
                <span className="repo-rev-sidebar-symbols-node__name e2e-symbol-name">{node.name}</span>
                {node.containerName && (
                    <span className="repo-rev-sidebar-symbols-node__container-name">
                        <small>{node.containerName}</small>
                    </span>
                )}
                <span className="repo-rev-sidebar-symbols-node__path">
                    <small>{node.location.resource.path}</small>
                </span>
            </NavLink>
        </li>
    )
}

class FilteredSymbolsConnection extends FilteredConnection<GQL.ISymbol, Pick<SymbolNodeProps, 'location'>> {}

interface Props {
    repoID: GQL.ID
    rev: string | undefined
    history: H.History
    location: H.Location
    /** The path of the file or directory currently shown in the content area */
    activePath: string
}

export class RepoRevSidebarSymbols extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()

    public componentDidUpdate(prevProps: Props): void {
        this.componentUpdates.next(this.props)
    }

    public render(): JSX.Element | null {
        return (
            <FilteredSymbolsConnection
                className="repo-rev-sidebar-symbols"
                compact={true}
                noun="symbol"
                pluralNoun="symbols"
                queryConnection={this.fetchSymbols}
                nodeComponent={SymbolNode}
                nodeComponentProps={{ location: this.props.location } as Pick<SymbolNodeProps, 'location'>}
                defaultFirst={100}
                useURLQuery={false}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private fetchSymbols = (args: { first?: number; query?: string }): Observable<GQL.ISymbolConnection> =>
        this.componentUpdates.pipe(
            startWith(this.props),
            map(props => ({ repoID: props.repoID, rev: props.rev, activePath: props.activePath })),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            switchMap(props =>
                fetchSymbols(props.repoID, props.rev || '', {
                    ...args,
                    // `includePatterns` expects regexes, so first escape the
                    // path.
                    includePatterns: [escapeRegExp(props.activePath)],
                })
            )
        )
}
