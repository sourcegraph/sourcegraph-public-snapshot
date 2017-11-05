import * as React from 'react'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/map'
import { Subscription } from 'rxjs/Subscription'
import { fetchRepoGroups } from './backend'

interface Props {
    /** The query value of the active search scope, or undefined if it's still loading */
    scopeQuery?: string
}

interface State {
    /**
     * A map from repogroup: name to a list of its member repositories.
     */
    repoGroups?: Map<string, string[]>
}

/**
 * The label showing the scope query.
 *
 * For repogroup: tokens in the query, it shows a hover listing the repositories
 * in the group.
 */
export class ScopeLabel extends React.Component<Props, State> {
    /**
     * Cache of the repo groups from the server. This will not be refreshed
     * until the user reloads the page.
     */
    private static REPO_GROUPS: Map<string, string[]> | undefined

    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        if (ScopeLabel.REPO_GROUPS) {
            this.state = {
                repoGroups: ScopeLabel.REPO_GROUPS,
            }
        } else {
            this.state = {}

            this.subscriptions.add(
                fetchRepoGroups()
                    .catch(err => {
                        console.error(err)
                        return []
                    })
                    .map((repoGroups: GQL.IRepoGroup[]) => {
                        const map = new Map<string, string[]>()
                        for (const { name, repositories } of repoGroups) {
                            map.set(name, repositories)
                        }
                        ScopeLabel.REPO_GROUPS = map // cache
                        return { repoGroups: map } as State
                    })
                    .subscribe(newState => this.setState(newState), err => console.error(err))
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const label = 'Scoped to: '
        let title: string | undefined

        const parts: React.ReactChild[] = []

        if (this.props.scopeQuery === '') {
            parts.push('Searching all code')
        } else {
            parts.push(label)
            if (this.props.scopeQuery === undefined) {
                parts.push('loading...')
            } else {
                title = `The scope query is merged with the user-provided query to perform the search`

                const tokens = parseQuery(this.props.scopeQuery)
                parts.push(
                    ...tokens.map((token, i) => {
                        if (typeof token === 'string') {
                            return (
                                <span key={i} className="scope-label2__query">
                                    {token}
                                </span>
                            )
                        }

                        // Show a list of repos that are members of this repogroup.
                        let title = ''
                        if (this.state.repoGroups && this.state.repoGroups.has(token.value)) {
                            title = `Repositories in group "${token.value}":\n${this.state.repoGroups.get(
                                token.value
                            )!.join('\n')}`
                        }
                        return (
                            <span key={i} className="scope-label2__query scope-label2__query-tip" title={title}>
                                {token.field}:{token.value}
                            </span>
                        )
                    })
                )
            }
        }

        return (
            <div className="scope-label2" title={title}>
                {joinWithWhitespaceTextNodes(parts)}
            </div>
        )
    }
}

interface RepoGroupToken {
    field: 'repogroup'
    value: string
}

/**
 * An intentionally simplified query parser that only specially extracts
 * repogroup: tokens so that their member repositories can be shown to
 * the user. It is best effort and must not be relied on for correctness.
 */
function parseQuery(query: string): (string | RepoGroupToken)[] {
    if (query.includes('"')) {
        // Preserve as a single unparsed token because our simple parser
        // would return incorrect results.
        return [query]
    }

    const tokens = query.split(/\s+/)
    return tokens.map(token => {
        if (token.startsWith('repogroup:')) {
            return { field: 'repogroup', value: token.slice('repogroup:'.length) } as RepoGroupToken
        }
        return token
    })
}

function joinWithWhitespaceTextNodes(elements: React.ReactChild[]): React.ReactChild[] {
    const elements2: React.ReactChild[] = []
    for (const e of elements) {
        elements2.push(e)
        elements2.push(' ')
    }
    return elements2
}
