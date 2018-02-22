import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { RepoLink } from '../repo/RepoLink'
import { eventLogger } from '../tracking/eventLogger'
import { ResultContainer } from './ResultContainer'

interface Props {
    /**
     * The repository that was matched.
     */
    result: GQL.IRepository

    /**
     * Called when the search result is selected.
     */
    onSelect: () => void
}

const logClickOnText = () => eventLogger.log('RepositorySearchResultClicked')

export const RepositorySearchResult: React.StatelessComponent<Props> = (props: Props) => (
    <ResultContainer
        titleClassName="repository-search-result__title"
        icon={RepoIcon}
        title={
            <>
                <RepoLink
                    repoPath={props.result.uri}
                    // tslint:disable-next-line:jsx-no-lambda
                    onClick={() => {
                        logClickOnText()
                        props.onSelect()
                    }}
                />
                <span className="repository-search-result__spacer" />
                <small>Repository name match</small>
            </>
        }
    />
)
