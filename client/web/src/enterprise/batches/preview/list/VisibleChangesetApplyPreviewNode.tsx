import React, { useCallback, useMemo, useState } from 'react'

import {
    mdiCardTextOutline,
    mdiFileDocumentEditOutline,
    mdiAccountEdit,
    mdiCheckboxBlankCircle,
    mdiChevronDown,
    mdiChevronRight,
} from '@mdi/js'
import classNames from 'classnames'

import type { Maybe } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Link, Alert, Icon, Tabs, TabList, TabPanels, TabPanel, Tab, H3, Tooltip } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../../components/diff/DiffStat'
import { InputTooltip } from '../../../../components/InputTooltip'
import { ChangesetState, type VisibleChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { Branch, BranchMerge } from '../../Branch'
import { Description } from '../../Description'
import { ChangesetStatusCell } from '../../detail/changesets/ChangesetStatusCell'
import { ExternalChangesetTitle } from '../../detail/changesets/ExternalChangesetTitle'
import type { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'
import { checkPublishability } from '../utils'

import type { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'
import { ChangesetSpecFileDiffConnection } from './ChangesetSpecFileDiffConnection'
import { GitBranchChangesetDescriptionInfo } from './GitBranchChangesetDescriptionInfo'
import { PreviewActions } from './PreviewActions'
import { PreviewNodeIndicator } from './PreviewNodeIndicator'

import styles from './VisibleChangesetApplyPreviewNode.module.scss'

export interface VisibleChangesetApplyPreviewNodeProps {
    node: VisibleChangesetApplyPreviewFields
    authenticatedUser: PreviewPageAuthenticatedUser
    selectable?: {
        onSelect: (id: string) => void
        isSelected: (id: string) => boolean
    }

    /** Used for testing. **/
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. **/
    expandChangesetDescriptions?: boolean
}

export const VisibleChangesetApplyPreviewNode: React.FunctionComponent<
    React.PropsWithChildren<VisibleChangesetApplyPreviewNodeProps>
> = ({ node, authenticatedUser, selectable, queryChangesetSpecFileDiffs, expandChangesetDescriptions = false }) => {
    const [isExpanded, setIsExpanded] = useState(expandChangesetDescriptions)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    return (
        <>
            <Button
                variant="icon"
                className="test-batches-expand-preview d-none d-sm-block mx-1"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />
            </Button>
            {selectable ? (
                <SelectBox node={node} selectable={selectable} />
            ) : (
                // 0-width empty element to allow us to keep the identical grid template of the parent
                // list, regardless of whether or not the nodes have the checkbox selector
                <span />
            )}
            <VisibleChangesetApplyPreviewNodeStatusCell
                node={node}
                className={classNames(
                    styles.visibleChangesetApplyPreviewNodeListCell,
                    styles.visibleChangesetApplyPreviewNodeCurrentState,
                    styles.visibleChangesetApplyPreviewNodeStatusCell,
                    'd-block d-sm-flex align-self-stretch'
                )}
            />
            <PreviewNodeIndicator node={node} />
            <PreviewActions
                node={node}
                className={classNames(
                    styles.visibleChangesetApplyPreviewNodeListCell,
                    styles.visibleChangesetApplyPreviewNodeAction,
                    'align-self-stretch'
                )}
            />
            <div
                className={classNames(
                    styles.visibleChangesetApplyPreviewNodeListCell,
                    styles.visibleChangesetApplyPreviewNodeInformation,
                    'align-self-stretch'
                )}
            >
                <div className="d-flex flex-column">
                    <ChangesetSpecTitle spec={node} />
                    <div className="mr-2">
                        <RepoLink spec={node} /> <References spec={node} />
                    </div>
                </div>
            </div>
            <div className="d-flex justify-content-center align-content-center align-self-stretch">
                {node.delta.commitMessageChanged && (
                    <div
                        className={classNames(
                            styles.visibleChangesetApplyPreviewNodeCommitChangeEntry,
                            'd-flex justify-content-center align-items-center flex-column mx-1'
                        )}
                    >
                        <Tooltip content="The commit message changed">
                            <Icon aria-label="The commit message changed" svgPath={mdiCardTextOutline} />
                        </Tooltip>
                        <span className="text-nowrap">Message</span>
                    </div>
                )}
                {node.delta.diffChanged && (
                    <div
                        className={classNames(
                            styles.visibleChangesetApplyPreviewNodeCommitChangeEntry,
                            'd-flex justify-content-center align-items-center flex-column mx-1'
                        )}
                    >
                        <Tooltip content="The diff for this changeset has been updated">
                            <Icon
                                aria-label="The diff for this changeset has been updated"
                                svgPath={mdiFileDocumentEditOutline}
                            />
                        </Tooltip>
                        <span className="text-nowrap" aria-hidden={true}>
                            Diff
                        </span>
                    </div>
                )}
                {(node.delta.authorNameChanged || node.delta.authorEmailChanged) && (
                    <div
                        className={classNames(
                            styles.visibleChangesetApplyPreviewNodeCommitChangeEntry,
                            'd-flex justify-content-center align-items-center flex-column mx-1'
                        )}
                    >
                        <Tooltip content="The commit author details changed">
                            <Icon aria-label="The commit author details changed" svgPath={mdiAccountEdit} />
                        </Tooltip>
                        <span className="text-nowrap">Author</span>
                    </div>
                )}
            </div>
            <div
                className={classNames(
                    styles.visibleChangesetApplyPreviewNodeListCell,
                    'd-flex justify-content-center align-items-center align-self-stretch'
                )}
            >
                <ApplyDiffStat spec={node} />
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <Button
                onClick={toggleIsExpanded}
                className={classNames(
                    styles.visibleChangesetApplyPreviewNodeShowDetails,
                    'd-block d-sm-none test-batches-expand-preview'
                )}
                outline={true}
                variant="secondary"
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </Button>
            {isExpanded && (
                <>
                    <div
                        className={classNames(
                            styles.visibleChangesetApplyPreviewNodeExpandedSection,
                            styles.visibleChangesetApplyPreviewNodeBgExpanded,
                            'pt-4'
                        )}
                    >
                        <ExpandedSection
                            node={node}
                            authenticatedUser={authenticatedUser}
                            queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                        />
                    </div>
                </>
            )}
        </>
    )
}

const SelectBox: React.FunctionComponent<
    React.PropsWithChildren<{
        node: VisibleChangesetApplyPreviewFields
        selectable: {
            onSelect: (id: string) => void
            isSelected: (id: string) => boolean
        }
    }>
> = ({ node, selectable }) => {
    const isPublishableResult = useMemo(() => checkPublishability(node), [node])

    const toggleSelected = useCallback((): void => {
        if (isPublishableResult.publishable) {
            selectable.onSelect(isPublishableResult.changesetSpecID)
        }
    }, [selectable, isPublishableResult])

    const input = isPublishableResult.publishable ? (
        // eslint-disable-next-line no-restricted-syntax
        <InputTooltip
            id={`select-changeset-${isPublishableResult.changesetSpecID}`}
            type="checkbox"
            checked={selectable.isSelected(isPublishableResult.changesetSpecID)}
            onChange={toggleSelected}
            tooltip="Click to select changeset for bulk-modifying the publication state"
            placement="right"
            aria-label="Click to select changeset for bulk-modifying the publication state"
        />
    ) : (
        // eslint-disable-next-line no-restricted-syntax
        <InputTooltip
            id="select-changeset-hidden"
            type="checkbox"
            checked={false}
            disabled={true}
            tooltip={isPublishableResult.reason}
            placement="right"
            aria-label={isPublishableResult.reason}
        />
    )

    return (
        <div className="d-flex p-2 align-items-center">
            {input}
            {isPublishableResult.publishable ? (
                <span className="pl-2 d-block d-sm-none text-nowrap">Modify publication state</span>
            ) : null}
        </div>
    )
}

const ExpandedSection: React.FunctionComponent<
    React.PropsWithChildren<{
        node: VisibleChangesetApplyPreviewFields
        authenticatedUser: PreviewPageAuthenticatedUser

        /** Used for testing. **/
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    }>
> = ({ node, authenticatedUser, queryChangesetSpecFileDiffs }) => {
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return (
            <Alert className="mb-0" variant="info">
                When run, the changeset <strong>{node.targets.changeset.title}</strong> in repo{' '}
                <strong>{node.targets.changeset.repository.name}</strong> will be removed from this batch change.
            </Alert>
        )
    }

    if (node.targets.changesetSpec.description.__typename === 'ExistingChangesetReference') {
        return (
            <Alert className="mb-0" variant="info">
                When run, the changeset with ID <strong>{node.targets.changesetSpec.description.externalID}</strong>{' '}
                will be imported from <strong>{node.targets.changesetSpec.description.baseRepository.name}</strong>.
            </Alert>
        )
    }

    return (
        <Tabs size="large">
            <TabList>
                <Tab>
                    <span className="text-content" data-tab-content="Changed files">
                        Changed files
                    </span>
                    {node.delta.diffChanged && (
                        <Tooltip content="Changes in this tab">
                            <small className="text-warning ml-2">
                                <Icon
                                    aria-label="Changes in this tab"
                                    className={styles.visibleChangesetApplyPreviewNodeChangeIndicator}
                                    svgPath={mdiCheckboxBlankCircle}
                                />
                            </small>
                        </Tooltip>
                    )}
                </Tab>

                <Tab>
                    <span className="text-content" data-tab-content="Description">
                        Description
                    </span>
                    {(node.delta.titleChanged || node.delta.bodyChanged) && (
                        <Tooltip content="Changes in this tab">
                            <small className="text-warning ml-2">
                                <Icon
                                    aria-label="Changes in this tab"
                                    className={styles.visibleChangesetApplyPreviewNodeChangeIndicator}
                                    svgPath={mdiCheckboxBlankCircle}
                                />
                            </small>
                        </Tooltip>
                    )}
                </Tab>

                <Tab>
                    <span className="text-content" data-tab-content="Commits">
                        Commits
                    </span>
                    {(node.delta.authorEmailChanged ||
                        node.delta.authorNameChanged ||
                        node.delta.commitMessageChanged) && (
                        <Tooltip content="Changes in this tab">
                            <small className="text-warning ml-2">
                                <Icon
                                    aria-label="Changes in this tab"
                                    className={styles.visibleChangesetApplyPreviewNodeChangeIndicator}
                                    svgPath={mdiCheckboxBlankCircle}
                                />
                            </small>
                        </Tooltip>
                    )}
                </Tab>
            </TabList>
            <TabPanels>
                <TabPanel className="pt-3">
                    {node.delta.diffChanged && (
                        <Alert variant="warning">
                            The files in this changeset have been altered from the previous version. These changes will
                            be pushed to the target branch.
                        </Alert>
                    )}
                    <ChangesetSpecFileDiffConnection
                        spec={node.targets.changesetSpec.id}
                        queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                    />
                </TabPanel>

                <TabPanel className="pt-3">
                    {node.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
                        node.delta.bodyChanged &&
                        node.targets.changeset.currentSpec?.description.__typename ===
                            'GitBranchChangesetDescription' && (
                            <>
                                <H3 className="text-muted">
                                    <del>{node.targets.changeset.currentSpec.description.title}</del>
                                </H3>
                                <del className="text-muted">
                                    <Description description={node.targets.changeset.currentSpec.description.body} />
                                </del>
                            </>
                        )}
                    <H3>
                        {node.targets.changesetSpec.description.title}{' '}
                        <small>
                            by{' '}
                            <PersonLink
                                person={
                                    node.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
                                    node.targets.changeset.author
                                        ? node.targets.changeset.author
                                        : {
                                              email:
                                                  authenticatedUser.emails.find(email => email.isPrimary)?.email || '',
                                              displayName: authenticatedUser.displayName || authenticatedUser.username,
                                              user: authenticatedUser,
                                          }
                                }
                            />
                        </small>
                    </H3>
                    <Description description={node.targets.changesetSpec.description.body} />
                </TabPanel>

                <TabPanel className="pt-3">
                    <GitBranchChangesetDescriptionInfo node={node} />
                </TabPanel>
            </TabPanels>
        </Tabs>
    )
}

const ChangesetSpecTitle: React.FunctionComponent<
    React.PropsWithChildren<{ spec: VisibleChangesetApplyPreviewFields }>
> = ({ spec }) => {
    // Identify the title and external ID/URL, if the changeset spec has them, depending on the type
    let externalID: Maybe<string> = null
    let externalURL: Maybe<{ url: string }> = null
    let title: Maybe<string> = null

    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsAttach') {
        // An import changeset does not display a regular title
        if (spec.targets.changesetSpec.description.__typename === 'ExistingChangesetReference') {
            return <H3>Import changeset #{spec.targets.changesetSpec.description.externalID}</H3>
        }

        title = spec.targets.changesetSpec.description.title
    } else {
        externalID = spec.targets.changeset.externalID
        externalURL = spec.targets.changeset.externalURL
        title = spec.targets.changeset.title
    }

    // For existing changesets, the title also may have been updated
    const newTitle =
        spec.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
        spec.targets.changesetSpec.description.__typename !== 'ExistingChangesetReference' &&
        spec.delta.titleChanged
            ? spec.targets.changesetSpec.description.title
            : null

    return (
        <H3>
            {newTitle ? (
                <>
                    <del className="mr-1">
                        <ExternalChangesetTitle
                            className="text-muted"
                            externalID={externalID}
                            externalURL={externalURL}
                            title={title}
                        />
                    </del>
                    {newTitle}
                </>
            ) : (
                <ExternalChangesetTitle externalID={externalID} externalURL={externalURL} title={title} />
            )}
        </H3>
    )
}

const RepoLink: React.FunctionComponent<React.PropsWithChildren<{ spec: VisibleChangesetApplyPreviewFields }>> = ({
    spec,
}) => {
    let to: string
    let name: string
    if (
        spec.targets.__typename === 'VisibleApplyPreviewTargetsAttach' ||
        spec.targets.__typename === 'VisibleApplyPreviewTargetsUpdate'
    ) {
        to = spec.targets.changesetSpec.description.baseRepository.url
        name = spec.targets.changesetSpec.description.baseRepository.name
    } else {
        to = spec.targets.changeset.repository.url
        name = spec.targets.changeset.repository.name
    }
    return (
        <Link to={to} target="_blank" rel="noopener noreferrer" className="d-block d-sm-inline">
            {name}
        </Link>
    )
}

const References: React.FunctionComponent<React.PropsWithChildren<{ spec: VisibleChangesetApplyPreviewFields }>> = ({
    spec,
}) => {
    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return null
    }
    if (spec.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    }
    return (
        <div className="d-block d-sm-inline-block">
            {spec.delta.baseRefChanged &&
                spec.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
                spec.targets.changeset.currentSpec?.description.__typename === 'GitBranchChangesetDescription' && (
                    <Branch
                        className="mr-2"
                        deleted={true}
                        name={spec.targets.changeset.currentSpec?.description.baseRef}
                    />
                )}
            <BranchMerge
                baseRef={spec.targets.changesetSpec.description.baseRef}
                forkTarget={spec.targets.changesetSpec.forkTarget}
                headRef={spec.targets.changesetSpec.description.headRef}
            />
        </div>
    )
}

const ApplyDiffStat: React.FunctionComponent<React.PropsWithChildren<{ spec: VisibleChangesetApplyPreviewFields }>> = ({
    spec,
}) => {
    let diffStat: { added: number; deleted: number }
    if (spec.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        if (!spec.targets.changeset.diffStat) {
            return null
        }
        diffStat = spec.targets.changeset.diffStat
    } else if (spec.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    } else {
        diffStat = spec.targets.changesetSpec.description.diffStat
    }
    return <DiffStatStack {...diffStat} />
}

const VisibleChangesetApplyPreviewNodeStatusCell: React.FunctionComponent<
    React.PropsWithChildren<Pick<VisibleChangesetApplyPreviewNodeProps, 'node'> & { className?: string }>
> = ({ node, className }) => {
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsAttach') {
        return <ChangesetStatusCell state={ChangesetState.UNPUBLISHED} className={className} />
    }
    return <ChangesetStatusCell state={node.targets.changeset.state} className={className} />
}
