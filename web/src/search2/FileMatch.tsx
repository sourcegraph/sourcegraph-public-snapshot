import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import { ReferencesGroup } from '../references/ReferencesWidget'
import { parseSearchURLQuery } from './index'

interface Props {
    location: H.Location

    /**
     * The file match search result.
     */
    result: GQL.IFileMatch

    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this file should be rendered as expanded.
     */
    expanded: boolean
}

export const FileMatch: React.StatelessComponent<Props> = (props: Props) => {
    const searchOptions = parseSearchURLQuery(props.location.search)
    const parsed = new URL(props.result.resource)
    const repoPath = parsed.hostname + parsed.pathname
    const rev = parsed.search.substr('?'.length)
    const filePath = parsed.hash.substr('#'.length)
    const refs = props.result.lineMatches.map(match => ({
        range: {
            start: {
                character: match.offsetAndLengths[0][0],
                line: match.lineNumber,
            },
            end: {
                character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
                line: match.lineNumber,
            },
        },
        uri: props.result.resource,
        repoURI: repoPath,
    }))

    return (
        <ReferencesGroup
            hidden={!props.expanded}
            repoPath={repoPath}
            localRev={rev}
            filePath={filePath}
            refs={refs}
            isLocal={false}
            icon={RepoIcon}
            onSelect={props.onSelect}
            searchOptions={searchOptions}
        />
    )
}
