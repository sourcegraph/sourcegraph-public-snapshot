import * as React from 'react'

import { Accordion } from '@reach/accordion'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { Alert, Button, H3, Icon, Link, Text } from '@sourcegraph/wildcard'

import { FetchOwnershipResult, FetchOwnershipVariables, SearchPatternType } from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'
import { FETCH_OWNERS } from './grapqlQueries'

import styles from './FileOwnershipPanel.module.scss'
import { storageKeyForPartial } from '../../../components/DismissibleAlert'
import { mdiClose } from '@mdi/js'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'

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
        return <div className={styles.contents}>Loading...</div>
    }

    if (error) {
        logger.log(error)
        return (
            <div className={styles.contents}>
                <Alert variant="danger">Error getting ownership data.</Alert>
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
            <Alert variant="info">No ownership data for this file.</Alert>
        </div>
    )
}

const OWN_EXPLANATION_KEY = 'own-explanation'

const OwnExplanation: React.FunctionComponent<{}> = () => {
    const [dismissed, setDismissed] = React.useState<boolean>(
        localStorage.getItem(storageKeyForPartial(OWN_EXPLANATION_KEY)) === 'true'
    )

    const onDismiss = React.useCallback(() => {
        localStorage.setItem(storageKeyForPartial(OWN_EXPLANATION_KEY), 'true')
        setDismissed(true)
    }, [])

    if (dismissed) {
        return null
    }

    return (
        <div className={classNames(styles.ownExplanation, 'd-flex align-items-start')}>
            <div className="flex-1">
                <H3>Sourcegraph Own Preview</H3>
                <Text className="mb-2">
                    Find code owners from a CODEOWNERS file in this repository, or from your external ownership tracking
                    system here. <Link to="/help/own">Learn more</Link>
                    <br />
                    In the future, we will suggest you many kinds of people to reach out to, including language experts,
                    codebase experts, and domain experts.
                </Text>
                <Text>Try Sourcegraph Own in Search as well!</Text>
                <Button variant="secondary" size="sm" outline={true} className="mr-2">
                    <SyntaxHighlightedSearchQuery
                        query="file:has.owner(johndoe)"
                        searchPatternType={SearchPatternType.standard}
                    />
                </Button>
                <Button variant="secondary" size="sm" outline={true}>
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
    )
}
