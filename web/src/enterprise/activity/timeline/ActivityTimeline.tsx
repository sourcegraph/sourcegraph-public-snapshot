import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useMemo, useState } from 'react'
import { first, take } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'

interface GQLIEvent {}

const queryEvents = async ({ query, extensionsController }: { query: string }): { nodes: GQLIEvent } =>
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
 * A chronological list of events with a header.
 */
export const ActivityTimeline: React.FunctionComponent<Props> = ({
    query,
    onQueryChange,
    extensionsController,
    ...props
}) => {
    const [eventsOrError, setEventsOrError] = useState<typeof LOADING | GQL.ISearchResults | ErrorLike>(LOADING)

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setEventsOrError(await queryEvents({ query, extensionsController }))
        } catch (err) {
            setEventsOrError(asError(err))
        }
    }, [extensionsController, query])

    return (
        <div className="events-list">
            {isErrorLike(eventsOrError) ? (
                <div className="alert alert-danger mt-2">{eventsOrError.message}</div>
            ) : (
                <>
                    <EventsListHeader {...props} query={query} onQueryChange={onQueryChange} />
                    {eventsOrError === LOADING ? (
                        <LoadingSpinner className="mt-2" />
                    ) : eventsOrError.resultCount === 0 ? (
                        <p className="p-2 mb-0 text-muted">No events found.</p>
                    ) : (
                        <ul className="list-group list-group-flush">
                            {eventsOrError.results.map((event, i) => (
                                <EventsListItem key={i} {...props} event={event as GQL.ICommitSearchResult} />
                            ))}
                        </ul>
                    )}
                </>
            )}
        </div>
    )
}
