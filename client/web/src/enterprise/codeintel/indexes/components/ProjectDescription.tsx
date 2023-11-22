import React, { type FunctionComponent } from 'react'

import { Badge, Link } from '@sourcegraph/wildcard'

import type { PreciseIndexFields } from '../../../../graphql-operations'

export interface ProjectDescriptionProps {
    index: PreciseIndexFields
}

export const ProjectDescription: FunctionComponent<ProjectDescriptionProps> = ({ index }) => (
    <>
        Directory <DirectoryDescription index={index} /> indexed at commit <CommitDescription index={index} /> by{' '}
        <IndexerDescription index={index} />
    </>
)

interface DirectoryDescriptionProps {
    index: PreciseIndexFields
}

const DirectoryDescription: FunctionComponent<DirectoryDescriptionProps> = ({ index }) =>
    index.projectRoot ? (
        <Link to={index.projectRoot.url}>
            <span>{index.projectRoot.path || '/'}</span>
        </Link>
    ) : (
        <span>{index.inputRoot || '/'}</span>
    )

interface CommitDescriptionProps {
    index: PreciseIndexFields
}

const CommitDescription: FunctionComponent<CommitDescriptionProps> = ({ index }) => (
    <>
        <Badge className="px-1 py-0" as="code">
            {index.projectRoot ? (
                <Link to={index.projectRoot.commit.url}>{index.projectRoot.commit.abbreviatedOID}</Link>
            ) : (
                <>{index.inputCommit.slice(0, 7)}</>
            )}
        </Badge>
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
}

const IndexerDescription: FunctionComponent<IndexerDescriptionProps> = ({ index }) => (
    <span>
        {index.indexer ? (
            index.indexer.url ? (
                <Link to={index.indexer.url}>{index.indexer.name}</Link>
            ) : (
                <>{index.indexer.name}</>
            )
        ) : (
            'an unknown indexer'
        )}
    </span>
)
