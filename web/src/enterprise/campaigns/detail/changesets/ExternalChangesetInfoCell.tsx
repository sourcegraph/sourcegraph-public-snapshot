import React from 'react'
import {
    ExternalChangesetFields,
    ChangesetExternalState,
    ChangesetPublicationState,
} from '../../../../graphql-operations'
import { Observer } from 'rxjs'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { ChangesetLabel } from './ChangesetLabel'
import { Link } from '../../../../../../shared/src/components/Link'
import { ChangesetLastSynced } from './ChangesetLastSynced'

export interface ExternalChangesetInfoCellProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    campaignUpdates?: Pick<Observer<void>, 'next'>
}

export const ExternalChangesetInfoCell: React.FunctionComponent<ExternalChangesetInfoCellProps> = ({
    node,
    viewerCanAdminister,
    campaignUpdates,
}) => (
    <div className="d-flex flex-column">
        <div className="m-0 mb-2">
            <h3 className="m-0 d-inline">
                <LinkOrSpan
                    /* Deleted changesets most likely don't exist on the codehost anymore and would return 404 pages */
                    to={
                        node.externalURL && node.externalState !== ChangesetExternalState.DELETED
                            ? node.externalURL.url
                            : undefined
                    }
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    {node.title}
                    {node.externalID && <>(#{node.externalID}) </>}
                    {node.externalURL && node.externalState !== ChangesetExternalState.DELETED && (
                        <>
                            {' '}
                            <ExternalLinkIcon size="1rem" />
                        </>
                    )}
                </LinkOrSpan>
            </h3>
            {node.labels.length > 0 && (
                <span className="ml-2">
                    {node.labels.map(label => (
                        <ChangesetLabel label={label} key={label.text} />
                    ))}
                </span>
            )}
        </div>
        <div>
            <strong className="mr-2">
                <Link to={node.repository.url} target="_blank" rel="noopener noreferrer">
                    {node.repository.name}
                </Link>
            </strong>
            {node.publicationState === ChangesetPublicationState.PUBLISHED && (
                <ChangesetLastSynced
                    changeset={node}
                    viewerCanAdminister={viewerCanAdminister}
                    campaignUpdates={campaignUpdates}
                />
            )}
        </div>
    </div>
)
