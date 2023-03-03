import * as React from 'react'

import { mdiClose } from '@mdi/js'
import { Accordion } from '@reach/accordion'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Alert, Button, ErrorAlert, H3, H4, Icon, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'
import { FetchOwnershipResult, FetchOwnershipVariables, SearchPatternType } from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'
import { FETCH_OWNERS } from './grapqlQueries'

import styles from './FileOwnershipPanel.module.scss'

export const FileOwnershipPanel: React.FunctionComponent<{
    repoID: string
    revision?: string
    filePath: string
}> = ({ repoID, revision, filePath }) => {
    const { data, loading, error } = useQuery<FetchOwnershipResult, FetchOwnershipVariables>(FETCH_OWNERS, {
        variables: {
            repo: repoID,
            revision: revision ?? '',
            currentPath: filePath,
        },
    })
    if (loading) {
        return (
            <div className={classNames(styles.loaderWrapper, 'text-muted')}>
                <LoadingSpinner inline={true} className="mr-1" /> Loading...
            </div>
        )
    }

    if (error) {
        logger.log(error)
        return (
            <div className={styles.contents}>
                <ErrorAlert error={error} prefix="Error getting ownership data" className="mt-2" />
            </div>
        )
    }

    if (
        data?.node &&
        data.node.__typename === 'Repository' &&
        data.node.commit?.blob &&
        data.node.commit.blob.ownership.nodes.length > 0
    ) {
        return (
            <>
                <OwnExplanation />
                <Accordion
                    as="table"
                    collapsible={true}
                    multiple={true}
                    className={classNames(styles.table, styles.contents)}
                >
                    <thead className="sr-only">
                        <tr>
                            <th>Show details</th>
                            <th>Contact</th>
                            <th>Owner</th>
                            <th>Reason</th>
                        </tr>
                    </thead>
                    {data.node.commit.blob?.ownership.nodes.map(ownership =>
                        ownership.owner.__typename === 'Person' ? (
                            <FileOwnershipEntry
                                key={ownership.owner.email}
                                person={ownership.owner}
                                reasons={ownership.reasons.filter(
                                    reason => reason.__typename === 'CodeownersFileEntry'
                                )}
                            />
                        ) : (
                            // TODO #48303: Add support for teams.
                            <></>
                        )
                    )}
                </Accordion>
            </>
        )
    }

    return (
        <div className={styles.contents}>
            <OwnExplanation />
            <Alert variant="info">No ownership data for this file.</Alert>
        </div>
    )
}

const OwnExplanation: React.FunctionComponent<{}> = () => {
    const [dismissed, setDismissed] = useTemporarySetting('own.panelExplanationHidden')

    const onDismiss = React.useCallback(() => {
        setDismissed(true)
    }, [setDismissed])

    if (dismissed) {
        return null
    }

    return (
        <MarketingBlock contentClassName={styles.ownExplanationContainer}>
            <div className="d-flex align-items-start">
                <div className="flex-1">
                    <H3 as={H4} className={styles.ownExplanationTitle}>
                        Sourcegraph Own Preview
                    </H3>
                    <Text className={classNames(styles.ownExplanationContent, 'mb-2')}>
                        Find code owners from a CODEOWNERS file in this repository, or from your external ownership
                        tracking system here. The <Link to="/help/own">Own documentation</Link> contains more
                        information.
                    </Text>
                    <Text className={classNames(styles.ownExplanationContent, 'mb-1')}>
                        Sourcegraph Own also works in search:
                    </Text>
                    <Button
                        variant="secondary"
                        size="sm"
                        outline={true}
                        as={Link}
                        to="/search?q=file:has.owner(johndoe)"
                        className="mr-2"
                    >
                        <SyntaxHighlightedSearchQuery
                            query="file:has.owner(johndoe)"
                            searchPatternType={SearchPatternType.standard}
                        />
                    </Button>
                    <Button variant="secondary" size="sm" as={Link} to="/search?q=select:file.owners" outline={true}>
                        <SyntaxHighlightedSearchQuery
                            query="select:file.owners"
                            searchPatternType={SearchPatternType.standard}
                        />
                    </Button>
                </div>
                <Button aria-label="Dismiss alert" variant="icon" onClick={onDismiss}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
        </MarketingBlock>
    )
}
