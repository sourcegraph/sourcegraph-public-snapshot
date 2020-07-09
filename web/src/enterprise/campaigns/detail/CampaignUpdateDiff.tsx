import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { HeroPage } from '../../../components/HeroPage'
import { ChangesetNode } from './changesets/ChangesetNode'
import { ThemeProps } from '../../../../../shared/src/theme'
import { HiddenExternalChangesetNode } from './changesets/HiddenExternalChangesetNode'
import { ChangesetStateIcon } from './changesets/ChangesetStateIcon'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { DiffStat } from '../../../components/diff/DiffStat'
import { Link } from '../../../../../shared/src/components/Link'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { Collapsible } from '../../../components/Collapsible'
import { ErrorMessage } from '../../../components/alerts'

interface Props extends ThemeProps {
    history: H.History
    location: H.Location
    campaignDelta: GQL.ICampaignDelta
    className?: string
}

/**
 * A list of a campaign's changesets changed over a new patch set
 */
export const CampaignUpdateDiff: React.FunctionComponent<Props> = ({
    campaignDelta,
    className,
    history,
    location,
    isLightTheme,
}) => {
    if (!campaignDelta.viewerCanAdminister) {
        return <HeroPage body="Updating a campaign is not permitted without campaign admin permissions." />
    }
    return (
        <div className={className}>
            <h3 className="mt-4 mb-2">Preview of changes</h3>
            <div className="alert alert-info mt-2">
                <AlertCircleIcon className="icon-inline" /> You are updating an existing campaign. By clicking 'Update',
                all above changesets that are not 'unmodified' will be updated on the codehost.
            </div>
            {campaignDelta.titleChanged && (
                <div className="form-group">
                    <h5>
                        <strong>Campaign title</strong>
                    </h5>
                    <span className="text-danger">- {campaignDelta.campaign.name}</span>
                    <span className="text-success">+ {campaignDelta.newTitle}</span>
                </div>
            )}
            {campaignDelta.descriptionChanged && (
                <div className="form-group">
                    <h5>
                        <strong>Campaign description</strong>
                    </h5>
                    <span className="text-danger">- {campaignDelta.campaign.description}</span>
                    <span className="text-success">+ {campaignDelta.newDescription}</span>
                </div>
            )}
            <h4>Changesets</h4>
            <ul className="list-group">
                {campaignDelta.changesets.edges.map((edge, index) =>
                    !edge.dirty ? (
                        <ChangesetNode
                            key={index}
                            node={edge.node}
                            history={history}
                            location={location}
                            isLightTheme={isLightTheme}
                            viewerCanAdminister={campaignDelta.viewerCanAdminister}
                        />
                    ) : (
                        <ChangesetUpdateEdge key={index} edge={edge} />
                    )
                )}
            </ul>
        </div>
    )
}

export const ChangesetUpdateEdge: React.FunctionComponent<{ edge: GQL.IChangesetUpdateEdge }> = ({ edge }) => {
    if (edge.node.__typename === 'HiddenExternalChangeset') {
        return <HiddenExternalChangesetNode node={edge.node} />
    }
    const changesetNodeRow = (
        <div className="d-flex align-items-start m-1 ml-2">
            <div className="changeset-node__content flex-fill">
                <div className="d-flex flex-column">
                    <div className="m-0 mb-2">
                        <h3 className="m-0 d-inline">
                            <ChangesetStateIcon state={edge.node.state} />
                            <LinkOrSpan
                                /* Deleted changesets most likely don't exist on the codehost anymore and would return 404 pages */
                                to={
                                    edge.node.externalURL && edge.node.state !== GQL.ChangesetState.DELETED
                                        ? edge.node.externalURL.url
                                        : undefined
                                }
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {edge.titleChanged && <s className="text-danger">{edge.node.title}</s>} {edge.newTitle}{' '}
                                {edge.node.externalID && <>(#{edge.node.externalID}) </>}
                                {edge.node.externalURL && edge.node.state !== GQL.ChangesetState.DELETED && (
                                    <ExternalLinkIcon size="1rem" />
                                )}
                                {edge.willPublish && <span className="badge mr-2 badge-success">New</span>}
                                {edge.willClose && <span className="badge mr-2 badge-danger">Will be closed</span>}
                                {edge.patchChanged && <span className="badge mr-2 badge-info">Updated diff</span>}
                                {edge.branchChanged && edge.node.state === GQL.ChangesetState.PENDING && (
                                    <span className="badge mr-2 badge-info">Branch name changed</span>
                                )}
                            </LinkOrSpan>
                        </h3>
                    </div>
                    <div>
                        <strong className="mr-2">
                            <Link to={edge.node.repository.url} target="_blank" rel="noopener noreferrer">
                                {edge.node.repository.name}
                            </Link>
                        </strong>
                        {edge.bodyChanged && <span className="text-warning">Changeset body changed.</span>}
                    </div>
                </div>
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                {edge.node.diffStat && <DiffStat {...edge.node.diffStat} expandedCounts={true} />}
            </div>
        </div>
    )
    return (
        <li className="list-group-item e2e-changeset-node">
            <Collapsible
                titleClassName="changeset-node__content flex-fill"
                expandedButtonClassName="mb-3"
                title={changesetNodeRow}
                wholeTitleClickable={false}
            >
                {/* {edge.node.diff?.fileDiffs && (
                    <FileDiffConnection
                        listClassName="list-group list-group-flush"
                        noun="changed file"
                        pluralNoun="changed files"
                        queryConnection={queryFileDiffs}
                        nodeComponent={FileDiffNode}
                        nodeComponentProps={{
                            history,
                            location,
                            isLightTheme,
                            persistLines: true,
                            extensionInfo: hydratedExtensionInfo,
                            lineNumbers: true,
                        }}
                        updateOnChange={node.repository.id}
                        defaultFirst={15}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        history={history}
                        location={location}
                        useURLQuery={false}
                        cursorPaging={true}
                    />
                )} */}
                {edge.node.error && (
                    <div className="alert alert-danger my-4">
                        <h3 className="alert-heading mb-0">Error while syncing changeset</h3>
                        <ErrorMessage error={edge.node.error} history={H.createMemoryHistory()} />
                    </div>
                )}
            </Collapsible>
        </li>
    )
}
