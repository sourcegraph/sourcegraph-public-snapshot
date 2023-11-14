import React, { useMemo } from 'react'

import classNames from 'classnames'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import type { BuildSearchQueryURLParameters, QueryState, SearchContextProps } from '@sourcegraph/shared/src/search'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendFilter, omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { getOwnerMatchUrl, type OwnerMatch } from '@sourcegraph/shared/src/search/stream'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link } from '@sourcegraph/wildcard'

import { ResultContainer } from './ResultContainer'

import styles from './OwnerSearchResult.module.scss'
import resultStyles from './SearchResult.module.scss'

export interface OwnerSearchResultProps
    extends Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        TelemetryProps,
        TelemetryV2Props {
    result: OwnerMatch
    onSelect: () => void
    containerClassName?: string
    as?: React.ElementType
    index: number

    // If not provided, the result will not contain a link to the owner's files.
    queryState?: QueryState
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
}

export const OwnerSearchResult: React.FunctionComponent<OwnerSearchResultProps> = ({
    result,
    onSelect,
    containerClassName,
    as,
    index,
    queryState,
    buildSearchURLQueryFromQueryState,
    selectedSearchContextSpec,
    telemetryService,
    telemetryRecorder,
}) => {
    const displayName = useMemo(() => {
        let displayName = ''
        if (result.type === 'person') {
            displayName =
                result.user?.displayName || result.user?.username || result.handle || result.email || 'Unknown person'
        } else if (result.type === 'team') {
            displayName = result.displayName || result.name || result.handle || result.email || 'Unknown team'
        }
        return displayName
    }, [result])

    const url = useMemo(() => {
        const url = getOwnerMatchUrl(result)
        const validUrlPrefixes = ['/teams/', '/users/', 'mailto:']
        // TODO(#54209): Introduce a proper solution where a streamed team
        // is returned with a URL if present. Temporarily return no URL
        // in case name contains /. This indicates a Github team, and these
        // are not linkable within code search - where / is not an allowed
        // character for team names.
        if (result.type === 'team' && result.name.includes('/')) {
            return ''
        }
        if (!validUrlPrefixes.some(prefix => url.startsWith(prefix))) {
            // This is not a real URL, remove it.
            return ''
        }
        return url
    }, [result])

    const fileSearchLink = useMemo(() => {
        if (!queryState || !buildSearchURLQueryFromQueryState) {
            return ''
        }

        const handle = result.handle || result.email
        if (!handle) {
            return ''
        }

        let query = queryState.query
        const selectFilter = findFilter(queryState.query, 'select', FilterKind.Global)
        if (selectFilter && selectFilter.value?.value === 'file.owners') {
            query = omitFilter(query, selectFilter)
        }
        query = appendFilter(query, 'file', `has.owner(${handle})`)

        const searchParams = buildSearchURLQueryFromQueryState({
            query,
            searchContextSpec: selectedSearchContextSpec,
        })
        return `/search?${searchParams}`
    }, [buildSearchURLQueryFromQueryState, queryState, result.email, result.handle, selectedSearchContextSpec])

    const logSearchOwnerClicked = (): void => {
        if (url.startsWith('mailto:')) {
            telemetryService.log('searchResults:ownershipMailto:clicked')
            telemetryRecorder.recordEvent('searchResults.ownershipMailto', 'clicked')
        }
        if (url.startsWith('/users/')) {
            telemetryService.log('searchResults:ownershipUsers:clicked')
            telemetryRecorder.recordEvent('searchResults.ownershipUsers', 'clicked')
        }
        if (url.startsWith('/teams/')) {
            telemetryService.log('searchResults:ownershipTeams:clicked')
            telemetryRecorder.recordEvent('searchResults.ownershipTeams', 'clicked')
        }
    }

    const title = (
        <div className="d-flex align-items-center">
            {result.type === 'person' ? (
                <UserAvatar
                    user={{
                        username: result.user?.username || displayName,
                        avatarURL: result.user?.avatarURL || null,
                        displayName,
                    }}
                    className={styles.avatar}
                    size={16}
                />
            ) : (
                <TeamAvatar
                    team={{ avatarURL: null, displayName: result.displayName || null, name: result.name }}
                    className={styles.avatar}
                    size={16}
                />
            )}

            {url ? (
                <Link to={url} className="text-muted" onClick={logSearchOwnerClicked}>
                    {displayName}
                </Link>
            ) : (
                <span className="text-muted">{displayName}</span>
            )}
        </div>
    )

    return (
        <ResultContainer
            className={containerClassName}
            as={as}
            index={index}
            title={title}
            detail={null}
            url="#"
            onClick={onSelect}
        >
            <div
                className={classNames(resultStyles.searchResultMatch, 'p-2 flex-column')}
                data-testid="owner-search-result"
            >
                <small className={resultStyles.matchType}>
                    <span>Owner match</span>
                    {fileSearchLink && (
                        <Link to={fileSearchLink} className={styles.filesLink}>
                            Show files
                        </Link>
                    )}
                </small>
                {result.type === 'person' && !result.user && (
                    <>
                        <div className={resultStyles.dividerVertical} />
                        <small className="d-block font-italic">
                            This owner is not associated with any Sourcegraph user or team.
                        </small>
                    </>
                )}
            </div>
        </ResultContainer>
    )
}
