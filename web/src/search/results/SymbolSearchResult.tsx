import * as React from 'react'
import { Link } from 'react-router-dom'
import { RepoFileLink } from '../../components/RepoFileLink'
import { ResultContainer } from '../../components/ResultContainer'
import { SymbolIcon } from '../../symbols/SymbolIcon'

interface SymbolSearchResultProps {
    result: GQL.ISymbol
}

export const SymbolSearchResult: React.StatelessComponent<SymbolSearchResultProps> = ({ result }) => (
    <ResultContainer
        collapsible={false}
        defaultExpanded={false}
        // tslint:disable-next-line:jsx-no-lambda
        icon={props => <SymbolIcon kind={result.kind} {...props} />}
        title={
            <>
                <RepoFileLink
                    filePath={result.location.resource.path}
                    repoPath={result.location.resource.repository.uri}
                />{' '}
                <code>
                    <Link to={result.url}>{result.name}</Link>{' '}
                    {result.containerName && <span className="text-muted">{result.containerName}</span>}
                </code>
            </>
        }
    />
)
