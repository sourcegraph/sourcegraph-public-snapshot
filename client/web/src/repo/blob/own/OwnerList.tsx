import * as React from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Alert, Button, ErrorAlert, H3, H4, Icon, Link, Text } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'
import { SearchPatternType, OwnerFields, OwnershipConnectionFields } from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'

import styles from './OwnerList.module.scss'

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
    isDirectory?: boolean
    makeOwnerButton?: (userId: string | undefined) => JSX.Element
    makeOwnerError?: Error
    repoID: string
    filePath: string
    refetch: any
}

export const OwnerList: React.FunctionComponent<OwnerListProps> = ({
    data,
    isDirectory = false,
    makeOwnerButton,
    makeOwnerError,
    repoID,
    filePath,
    refetch,
}) => {
    const [removeOwnerError, setRemoveOwnerError] = React.useState<Error | undefined>(undefined)

    if (data?.nodes && data.nodes.length) {
        const nodes = data.nodes
        const totalCount = data.totalOwners
        return (
            <div className={styles.contents}>
                <OwnExplanation owners={nodes.map(ownership => ownership.owner)} />
                {makeOwnerError && (
                    <div className={styles.contents}>
                        <ErrorAlert error={makeOwnerError} prefix="Error promoting an owner" className="mt-2" />
                    </div>
                )}
                {removeOwnerError && (
                    <div className={styles.contents}>
                        <ErrorAlert error={removeOwnerError} prefix="Error promoting an owner" className="mt-2" />
                    </div>
                )}
                <table className={styles.table}>
                    <thead>
                        <tr className="sr-only">
                            <th>Contact</th>
                            <th>Owner</th>
                            <th>Reason</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <th colSpan={3}>
                                {totalCount === 0 ? (
                                    <NoOwnershipAlert isDirectory={isDirectory} />
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
                                <React.Fragment key={index}>
                                    {index > 0 && <tr className={styles.bordered} />}
                                    <FileOwnershipEntry
                                        refetch={refetch}
                                        owner={ownership.owner}
                                        repoID={repoID}
                                        filePath={filePath}
                                        reasons={ownership.reasons}
                                        setRemoveOwnerError={setRemoveOwnerError}
                                    />
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
                            .map((ownership, index) => {
                                const userId =
                                    ownership.owner.__typename === 'Person' &&
                                    ownership.owner.user?.__typename === 'User'
                                        ? ownership.owner.user.id
                                        : undefined

                                return (
                                    <React.Fragment key={index}>
                                        {index > 0 && <tr className={styles.bordered} />}
                                        <FileOwnershipEntry
                                            owner={ownership.owner}
                                            reasons={ownership.reasons}
                                            makeOwnerButton={makeOwnerButton?.(userId)}
                                            repoID={repoID}
                                            filePath={filePath}
                                            refetch={refetch}
                                            setRemoveOwnerError={setRemoveOwnerError}
                                        />
                                    </React.Fragment>
                                )
                            })}
                    </tbody>
                </table>
            </div>
        )
    }

    return (
        <div className={styles.contents}>
            <OwnExplanation />
            <NoOwnershipAlert isDirectory={isDirectory} />
        </div>
    )
}

const NoOwnershipAlert: React.FunctionComponent<{ isDirectory?: boolean }> = ({ isDirectory }) => (
    <Alert variant="info">
        {isDirectory ? 'No ownership data for this path.' : 'No ownership data for this file.'}
    </Alert>
)
