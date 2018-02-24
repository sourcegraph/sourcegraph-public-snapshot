import KindIcon from '@sourcegraph/icons/lib/Kind'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { RepoFileLink } from '../components/RepoFileLink'
import { ResultContainer } from './ResultContainer'

interface SymbolSearchResultProps {
    result: GQL.ISymbol
}

export const SymbolSearchResult: React.StatelessComponent<SymbolSearchResultProps> = props => (
    <ResultContainer
        collapsible={false}
        defaultExpanded={false}
        icon={KindIcon}
        title={
            <>
                <RepoFileLink
                    filePath={props.result.location.resource.path}
                    repoPath={props.result.location.resource.repository.uri}
                />{' '}
                <code>
                    <Link to={props.result.url}>{props.result.name}</Link>{' '}
                    {props.result.containerName && <span className="text-muted">{props.result.containerName}</span>}
                </code>
            </>
        }
    />
)
