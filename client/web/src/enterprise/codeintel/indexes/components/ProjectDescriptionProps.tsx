import { FunctionComponent } from 'react'

import { Badge, Code, Link } from '@sourcegraph/wildcard'

import { PreciseIndexFields } from '../../../../graphql-operations'

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
            <strong>{index.projectRoot.path || '/'}</strong>
        </Link>
    ) : (
        <span>{index.inputRoot || '/'}</span>
    )

interface CommitDescriptionProps {
    index: PreciseIndexFields
}

const CommitDescription: FunctionComponent<CommitDescriptionProps> = ({ index }) => (
    <>
        <Code>
            {index.projectRoot ? (
                <Link to={index.projectRoot.commit.url}>
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
}

const IndexerDescription: FunctionComponent<IndexerDescriptionProps> = ({ index }) => (
    <span>
        {index.indexer &&
            (index.indexer.url === '' ? (
                <>{index.indexer.name}</>
            ) : (
                <Link to={index.indexer.url}>{index.indexer.name}</Link>
            ))}
    </span>
)
