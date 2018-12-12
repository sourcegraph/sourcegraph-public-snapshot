import * as React from 'react'
import { RepositoryIcon } from '../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import { RepoLink } from '../../../../shared/src/components/RepoLink'
import { ResultContainer } from '../../../../shared/src/components/ResultContainer'
import * as GQL from '../../../../shared/src/graphql/schema'
import { eventLogger } from '../../tracking/eventLogger'

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

export const RepositorySearchResult: React.FunctionComponent<Props> = (props: Props) => (
    <ResultContainer
        titleClassName="repository-search-result__title"
        icon={RepositoryIcon}
        title={
            <>
                <RepoLink
                    repoName={props.result.name}
                    to={props.result.url}
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
