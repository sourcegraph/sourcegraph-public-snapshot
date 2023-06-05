import * as React from 'react'
import { useEffect } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Button, ErrorAlert, H3, H4, Icon, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'
import {
    FetchOwnershipResult,
    FetchOwnershipVariables,
    OwnerFields,
    OwnershipConnectionFields,
    SearchPatternType,
} from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'
import { FETCH_OWNERS } from './grapqlQueries'

import styles from './FileOwnershipPanel.module.scss'

export const FileOwnershipPanel: React.FunctionComponent<
    {
        repoID: string
        revision?: string
        filePath: string
    } & TelemetryProps
> = ({ repoID, revision, filePath, telemetryService }) => {
    useEffect(() => {
        telemetryService.log('OwnershipPanelOpened')
    }, [telemetryService])

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

    if (data?.node?.__typename === 'Repository') {
        return <OwnerList data={data?.node?.commit?.blob?.ownership} />
    }
    return <OwnerList />
}

interface OwnExplanationProps {
    owners?: OwnerFields[]
}

const OwnExplanation: React.FunctionComponent<OwnExplanationProps> = ({ owners }) => {
    const [dismissed, setDismissed] = useTemporarySetting('own.panelExplanationHidden')

    const onDismiss = React.useCallback(() => {
        setDismissed(true)
    }, [setDismissed])

    if (dismissed) {
        return null
    }

    const ownerSearchPredicate = resolveOwnerSearchPredicate(owners)

    return (
        <MarketingBlock contentClassName={styles.ownExplanationContainer} wrapperClassName="mb-3">
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
                        to={`/search?q=file:has.owner(${ownerSearchPredicate})`}
                        className="mr-2"
                    >
                        <SyntaxHighlightedSearchQuery
                            query={`file:has.owner(${ownerSearchPredicate})`}
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

const resolveOwnerSearchPredicate = (owners?: OwnerFields[]): string => {
    if (owners) {
        for (const owner of owners) {
            if (owner.__typename === 'Person' && owner.user?.username) {
                return `@${owner.user.username}`
            }
        }
    }
    return 'johndoe'
}

interface OwnerListProps {
    data?: OwnershipConnectionFields
}

const OwnerList: React.FunctionComponent<OwnerListProps> = ({ data }) => {
    if (data?.nodes && data.nodes.length) {
        const nodes = data.nodes
        const totalCount = data.totalOwners
        return (
            <div className={styles.contents}>
                <OwnExplanation owners={nodes.map(ownership => ownership.owner)} />
                <table className={styles.table}>
                    <thead>
                        <tr className="sr-only">
                            <th>Contact</th>
                            <th>Owner</th>
                            <th>Reason</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <th colSpan={3}>
                                {totalCount === 0 ? (
                                    <Alert variant="info">No ownership data for this file.</Alert>
                                ) : (
                                    <H4 className="mb-3">Owners</H4>
                                )}
                            </th>
                        </tr>
                        {nodes
                            .filter(ownership =>
                                ownership.reasons.some(
                                    reason =>
                                        reason.__typename === 'CodeownersFileEntry' ||
                                        reason.__typename === 'AssignedOwner'
                                )
                            )
                            .map((ownership, index) => (
                                // This list is not expected to change, so it's safe to use the index as a key.
                                // eslint-disable-next-line react/no-array-index-key
                                <React.Fragment key={index}>
                                    {index > 0 && <tr className={styles.bordered} />}
                                    <FileOwnershipEntry owner={ownership.owner} reasons={ownership.reasons} />
                                </React.Fragment>
                            ))}
                        {
                            /* Visually separate two sets with a horizontal rule (like subsequent owners are)
                             * if there is data in both owners and signals.
                             */
                            totalCount > 0 && nodes.length > totalCount && <tr className={styles.bordered} />
                        }
                        {nodes.length > totalCount && (
                            <tr>
                                <th colSpan={3}>
                                    <H4 className="mt-3 mb-2">Inference signals</H4>
                                    <Text className={styles.ownInferenceExplanation}>
                                        These users have viewed or contributed to the file but are not registered owners
                                        of the file.
                                    </Text>
                                </th>
                            </tr>
                        )}
                        {nodes
                            .filter(
                                ownership =>
                                    !ownership.reasons.some(
                                        reason =>
                                            reason.__typename === 'CodeownersFileEntry' ||
                                            reason.__typename === 'AssignedOwner'
                                    )
                            )
                            .map((ownership, index) => (
                                // This list is not expected to change, so it's safe to use the index as a key.
                                // eslint-disable-next-line react/no-array-index-key
                                <React.Fragment key={index}>
                                    {index > 0 && <tr className={styles.bordered} />}
                                    <FileOwnershipEntry owner={ownership.owner} reasons={ownership.reasons} />
                                </React.Fragment>
                            ))}
                    </tbody>
                </table>
            </div>
        )
    }

    return (
        <div className={styles.contents}>
            <OwnExplanation />
            <Alert variant="info">No ownership data for this file.</Alert>
        </div>
    )
}
