import { RouteComponentProps } from 'react-router';
import { makeRepoURI, ParsedRepoURI } from 'sourcegraph/repo';
import { parseHash } from 'sourcegraph/util/url';

export interface ParsedRouteProps extends Partial<ParsedRepoURI>, RouteComponentProps<string[]> {
    routeName?: string;
    uri?: string;
}

export function parseRouteProps(props: RouteComponentProps<string[]>): ParsedRouteProps {
    if (!props.match.params[0]) {
        return props;
    }
    if (props.match.params[0] === 'search') {
        return { ...props, routeName: 'search' };
    }

    const uriPathSplit = props.match.params[0].split('/-/');
    const repoRevSplit = uriPathSplit[0].split('@');
    const hash = parseHash(props.location.hash);
    const position = hash.line ? { line: hash.line, char: hash.char } : undefined;
    const repoParams = { ...props, routeName: 'repository', repoPath: repoRevSplit[0], rev: repoRevSplit[1], position };
    if (uriPathSplit.length === 1) {
        return {...repoParams, uri: makeRepoURI(repoParams)};
    }
    const filePath = uriPathSplit[1].split('/').slice(1).join('/');
    const repoParamsWithPath = { ...repoParams, filePath };
    return {...repoParamsWithPath, uri: makeRepoURI(repoParams)};
}
