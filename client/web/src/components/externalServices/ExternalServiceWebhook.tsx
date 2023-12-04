import React from 'react'

import { Alert, Link, H3, Text, H4 } from '@sourcegraph/wildcard'

import { type ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'
import { CopyableText } from '../CopyableText'

interface Props {
    externalService: Pick<ExternalServiceFields, 'kind' | 'webhookURL'>
    className?: string
}

export const ExternalServiceWebhook: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalService: { kind, webhookURL },
    className,
}) => {
    if (!webhookURL) {
        return <></>
    }

    let description = <Text />

    switch (kind) {
        case ExternalServiceKind.BITBUCKETSERVER: {
            description = (
                <Text>
                    <Link
                        to="/help/admin/external_service/bitbucket_server#webhooks"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Webhooks
                    </Link>{' '}
                    will be created automatically on the configured Bitbucket Server instance. In case you don't provide
                    an admin token,{' '}
                    <Link
                        to="/help/admin/external_service/bitbucket_server#manual-configuration"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        follow the docs on how to set up webhooks manually
                    </Link>
                    .
                    <br />
                    To set up another webhook manually, use the following URL:
                </Text>
            )
            break
        }

        case ExternalServiceKind.GITHUB: {
            description = commonDescription('github')
            break
        }

        case ExternalServiceKind.GITLAB: {
            description = commonDescription('gitlab')
            break
        }
    }

    return (
        <Alert variant="info" className={className}>
            <H3>Batch changes webhooks</H3>
            <H4>
                Adding webhooks via code host connections has been{' '}
                <Link
                    to="/help/admin/config/webhooks/incoming#deprecation-notice"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    deprecated.
                </Link>
            </H4>
            {description}
            <CopyableText className="mb-2" text={webhookURL} size={webhookURL.length} />
            <Text className="mb-0">
                Note that only{' '}
                <Link to="/help/batch_changes" target="_blank" rel="noopener noreferrer">
                    batch changes
                </Link>{' '}
                make use of this webhook. To enable webhooks to trigger repository updates on Sourcegraph,{' '}
                <Link to="/help/admin/repo/webhooks" target="_blank" rel="noopener noreferrer">
                    see the docs on how to use them
                </Link>
                .
            </Text>
        </Alert>
    )
}

function commonDescription(url: string): JSX.Element {
    return (
        <Text>
            Point{' '}
            <Link to={`/help/admin/external_service/${url}#webhooks`} target="_blank" rel="noopener noreferrer">
                webhooks
            </Link>{' '}
            for this code host connection at the following URL:
        </Text>
    )
}
