import { useEffect, type FunctionComponent } from 'react'

import { mdiMagnify } from '@mdi/js'
import classNames from 'classnames'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useQuery } from '@sourcegraph/http-client'
import { useSettingsCascade } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import {
    Alert,
    Button,
    Container,
    ErrorAlert,
    H3,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    Text,
} from '@sourcegraph/wildcard'

import type {
    SavedSearchFields,
    SavedSearchResult,
    SavedSearchVariables,
    SearchPatternType,
} from '../graphql-operations'
import { LibraryItemStatusBadge, LibraryItemVisibilityBadge } from '../library/itemBadges'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { defaultPatternTypeFromSettings } from '../util/settings'

import { SAVED_SEARCH_UPDATED_LOCATION_STATE_KEY } from './EditPage'
import { savedSearchQuery } from './graphql'
import { SavedSearchPage } from './Page'
import { telemetryRecordSavedSearchViewSearchResults } from './telemetry'
import { urlToEditSavedSearch } from './util'

import styles from './DetailPage.module.scss'

/**
 * Page to show a saved search.
 */
export const DetailPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    const { id } = useParams<{ id: string }>()

    const { data, loading, error } = useQuery<SavedSearchResult, SavedSearchVariables>(savedSearchQuery, {
        variables: { id: id! },
    })
    const savedSearch = data?.node?.__typename === 'SavedSearch' ? data.node : null

    // Flash after updating.
    const location = useLocation()
    const navigate = useNavigate()
    const justUpdated = !!location.state?.[SAVED_SEARCH_UPDATED_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justUpdated) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justUpdated, navigate])
    const flash = justUpdated ? 'Saved!' : null

    return (
        <SavedSearchPage
            title={savedSearch ? `${savedSearch.description} - saved search` : 'Saved search'}
            actions={
                savedSearch?.viewerCanAdminister && (
                    <Button to={urlToEditSavedSearch(savedSearch)} variant="secondary" as={Link}>
                        Edit
                    </Button>
                )
            }
            breadcrumbsNamespace={savedSearch?.owner}
            breadcrumbs={savedSearch ? <PageHeader.Breadcrumb>{savedSearch.description}</PageHeader.Breadcrumb> : null}
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !savedSearch ? (
                <Alert variant="danger" as="p">
                    Saved search not found.
                </Alert>
            ) : (
                <>
                    <Detail savedSearch={savedSearch} telemetryRecorder={telemetryRecorder} />
                    {flash && !loading && (
                        <Alert variant="success" className="my-3">
                            {flash}
                        </Alert>
                    )}
                </>
            )}
        </SavedSearchPage>
    )
}

const Detail: FunctionComponent<TelemetryV2Props & { savedSearch: SavedSearchFields }> = ({
    savedSearch,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('savedSearches.detail', 'view', {
            metadata: namespaceTelemetryMetadata(savedSearch.owner),
        })
    }, [telemetryRecorder, savedSearch.owner])

    const defaultPatternType: SearchPatternType = defaultPatternTypeFromSettings(useSettingsCascade())
    const searchURL = `/search?${buildSearchURLQuery(savedSearch.query, defaultPatternType, false)}`
    return (
        <>
            <Text>
                <LibraryItemVisibilityBadge item={savedSearch} className="mr-1" />
                <LibraryItemStatusBadge item={savedSearch} className="mr-1" />
                <small>
                    Last updated <Timestamp date={savedSearch.updatedAt} noAbout={true} />
                    {savedSearch.updatedBy && (
                        <>
                            {' '}
                            by{' '}
                            <Link to={savedSearch.updatedBy.url}>
                                <strong>{savedSearch.updatedBy.username}</strong>
                            </Link>
                        </>
                    )}
                </small>
            </Text>
            <Container className={classNames(styles.container)}>
                <div className="d-flex flex-column flex-gap-2 align-items-center">
                    <H3>{savedSearch.description}</H3>
                    <SyntaxHighlightedSearchQuery query={savedSearch.query} />
                </div>
                <Button
                    variant="primary"
                    size="lg"
                    to={searchURL}
                    as={Link}
                    onClick={() =>
                        telemetryRecordSavedSearchViewSearchResults(telemetryRecorder, savedSearch, 'Detail')
                    }
                >
                    <Icon aria-hidden={true} svgPath={mdiMagnify} className="flex-shrink-0" size="sm" /> Run search
                </Button>
            </Container>
        </>
    )
}
