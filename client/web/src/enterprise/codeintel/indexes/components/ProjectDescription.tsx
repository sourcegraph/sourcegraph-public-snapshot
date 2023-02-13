import React, { FunctionComponent } from 'react'

import { Badge, Code, Link } from '@sourcegraph/wildcard'

import { PreciseIndexFields } from '../../../../graphql-operations'

export interface ProjectDescriptionProps {
    index: PreciseIndexFields
    onLinkClick?: React.MouseEventHandler
}

export const ProjectDescription: FunctionComponent<ProjectDescriptionProps> = ({ index, onLinkClick }) => (
    <>
        Directory <DirectoryDescription index={index} onLinkClick={onLinkClick} /> indexed at commit{' '}
        <CommitDescription index={index} onLinkClick={onLinkClick} /> by{' '}
        <IndexerDescription index={index} onLinkClick={onLinkClick} />
    </>
)

interface DirectoryDescriptionProps {
    index: PreciseIndexFields
    onLinkClick?: React.MouseEventHandler
}

const DirectoryDescription: FunctionComponent<DirectoryDescriptionProps> = ({ index, onLinkClick }) =>
    index.projectRoot ? (
        <Link to={index.projectRoot.url} onClick={onLinkClick}>
            <strong>{index.projectRoot.path || '/'}</strong>
        </Link>
    ) : (
        <span>{index.inputRoot || '/'}</span>
    )

interface CommitDescriptionProps {
    index: PreciseIndexFields
    onLinkClick?: React.MouseEventHandler
}

const CommitDescription: FunctionComponent<CommitDescriptionProps> = ({ index, onLinkClick }) => (
    <>
        <Code>
            {index.projectRoot ? (
                <Link to={index.projectRoot.commit.url} onClick={onLinkClick}>
                    <Code>{index.projectRoot.commit.abbreviatedOID}</Code>
                </Link>
            ) : (
                <span>{index.inputCommit.slice(0, 7)}</span>
            )}
        </Code>
        {index.tags.length > 0 && (
            <>
                {' '}
                (tagged as{' '}
                {index.tags
                    .slice(0, 3)
                    .map<React.ReactNode>(tag => (
                        <Badge key={tag} variant="outlineSecondary">
                            {tag}
                        </Badge>
                    ))
                    .reduce((previous, current) => [previous, ', ', current])}
                {index.tags.length > 3 && <> and {index.tags.length - 3} more</>})
            </>
        )}
    </>
)

interface IndexerDescriptionProps {
    index: PreciseIndexFields
    onLinkClick?: React.MouseEventHandler
}

const IndexerDescription: FunctionComponent<IndexerDescriptionProps> = ({ index, onLinkClick }) => (
    <span>
        {index.indexer ? (
            index.indexer.url ? (
                <Link to={index.indexer.url} onClick={onLinkClick}>
                    {index.indexer.name}
                </Link>
            ) : (
                <>{index.indexer.name}</>
            )
        ) : (
            'an unknown indexer'
        )}
    </span>
)
