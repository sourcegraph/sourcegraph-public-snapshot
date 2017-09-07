import { RouteComponentProps } from 'react-router'
import { makeRepoURI, ParsedRepoURI } from 'sourcegraph/repo'
import { parseHash } from 'sourcegraph/util/url'

export interface ParsedRouteProps extends Partial<ParsedRepoURI>, RouteComponentProps<any> { // the typed parameters are not useful in the parsed props, as you shouldn't use them
    routeName?: 'home' | 'search' | 'repository'
    uri?: string
}

export function parseRouteProps<T extends string | {[key: string]: string} | string[]>(props: RouteComponentProps<T>): ParsedRouteProps {
    if (props.location.pathname === '/') {
        return { ...props, routeName: 'home' }
    }
    if (props.location.pathname === '/search') {
        return { ...props, routeName: 'search' }
    }

    const uriPathSplit = props.match.params[0].split('/-/')
    const repoRevSplit = uriPathSplit[0].split('@')
    const hash = parseHash(props.location.hash)
    const position = hash.line ? { line: hash.line, char: hash.char } : undefined
    const repoParams = { ...props, routeName: 'repository' as 'repository', repoPath: repoRevSplit[0], rev: repoRevSplit[1], position }
    if (uriPathSplit.length === 1) {
        return {...repoParams, uri: makeRepoURI(repoParams)}
    }
    const filePath = uriPathSplit[1].split('/').slice(1).join('/')
    const repoParamsWithPath = { ...repoParams, filePath }
    return {...repoParamsWithPath, uri: makeRepoURI(repoParams)}
}
