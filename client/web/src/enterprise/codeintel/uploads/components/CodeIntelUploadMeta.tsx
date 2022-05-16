import { FunctionComponent } from 'react'

import { CardSubtitle, CardText, CardTitle, CardBody, Card } from '@sourcegraph/wildcard'

import { LsifUploadFields } from '../../../../graphql-operations'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../../shared/components/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'

export interface CodeIntelUploadMetaProps {
    node: LsifUploadFields
    now?: () => Date
}

export const CodeIntelUploadMeta: FunctionComponent<React.PropsWithChildren<CodeIntelUploadMetaProps>> = ({
    node,
    now,
}) => (
    <Card>
        <CardBody>
            <CardTitle>
                <CodeIntelUploadOrIndexRepository node={node} />
            </CardTitle>

            <CardSubtitle className="mb-2 text-muted">
                <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
            </CardSubtitle>

            <CardText>
                Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
            </CardText>
        </CardBody>
    </Card>
)
