import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useMemo, useState } from 'react'
import { first, take } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { fetchSearchResultStats, search } from '../../../search/backend'
import { ChangesListHeader } from './ChangesListHeader'
import { ChangesListItem } from './ChangesListItem'

const queryChanges = async ({
    query,
    extensionsController,
}: { query: string } & ExtensionsControllerProps<'services'>) =>
    search(`type:diff repo:codeintel|go-diff|csp ${query}`, { extensionsController })
        .pipe(first())
        .toPromise()

interface Props extends ExtensionsControllerProps<'services'>, QueryParameterProps {
    history: H.History
    location: H.Location
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The list of changes with a header.
 */
export const ChangesList: React.FunctionComponent<Props> = ({
    query,
    onQueryChange,
    extensionsController,
    ...props
}) => {
    const [changesOrError, setChangesOrError] = useState<typeof LOADING | GQL.ISearchResults | ErrorLike>(LOADING)

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setChangesOrError(await queryChanges({ query, extensionsController }))
        } catch (err) {
            setChangesOrError(asError(err))
        }
    }, [extensionsController, query])

    return (
        <div className="changes-list">
            {isErrorLike(changesOrError) ? (
                <div className="alert alert-danger mt-2">{changesOrError.message}</div>
            ) : (
                <>
                    <ChangesListHeader {...props} query={query} onQueryChange={onQueryChange} />
                    {changesOrError === LOADING ? (
                        <LoadingSpinner className="mt-2" />
                    ) : changesOrError.resultCount === 0 ? (
                        <p className="p-2 mb-0 text-muted">No changes found.</p>
                    ) : (
                        <ul className="list-group list-group-flush">
                            {changesOrError.results.map((change, i) => (
                                <ChangesListItem key={i} {...props} change={change as GQL.ICommitSearchResult} />
                            ))}
                        </ul>
                    )}
                </>
            )}
        </div>
    )
}
