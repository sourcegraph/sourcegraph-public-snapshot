import React from 'react'

import { Alert, Link, Typography } from '@sourcegraph/wildcard'

import { ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'
import { CopyableText } from '../CopyableText'

interface Props {
    externalService: Pick<ExternalServiceFields, 'kind' | 'webhookURL'>
}

export const ExternalServiceWebhook: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalService: { kind, webhookURL },
}) => {
    if (!webhookURL) {
        return <></>
    }

    let description = <p />

    switch (kind) {
        case ExternalServiceKind.BITBUCKETSERVER:
            description = (
                <p>
                    <Link
                        to="https://docs.sourcegraph.com/admin/external_service/bitbucket_server#webhooks"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Webhooks
                    </Link>{' '}
                    will be created automatically on the configured Bitbucket Server instance. In case you don't provide
                    an admin token,{' '}
                    <Link
                        to="https://docs.sourcegraph.com/admin/external_service/bitbucket_server#manual-configuration"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        follow the docs on how to set up webhooks manually
                    </Link>
                    .
                    <br />
                    To set up another webhook manually, use the following URL:
                </p>
            )
            break

        case ExternalServiceKind.GITHUB:
            description = commonDescription('github')
            break

        case ExternalServiceKind.GITLAB:
            description = commonDescription('gitlab')
            break
    }

    return (
        <Alert variant="info">
            <Typography.H3>Batch changes webhooks</Typography.H3>
            {description}
            <CopyableText className="mb-2" text={webhookURL} size={webhookURL.length} />
            <p className="mb-0">
                Note that only{' '}
                <Link to="https://docs.sourcegraph.com/user/batch_changes" target="_blank" rel="noopener noreferrer">
                    batch changes
                </Link>{' '}
                make use of this webhook. To enable webhooks to trigger repository updates on Sourcegraph,{' '}
                <Link to="https://docs.sourcegraph.com/admin/repo/webhooks" target="_blank" rel="noopener noreferrer">
                    see the docs on how to use them
                </Link>
                .
            </p>
        </Alert>
    )
}

function commonDescription(url: string): JSX.Element {
    return (
        <p>
            Point{' '}
            <Link
                to={`https://docs.sourcegraph.com/admin/external_service/${url}#webhooks`}
                target="_blank"
                rel="noopener noreferrer"
            >
                webhooks
            </Link>{' '}
            for this code host connection at the following URL:
        </p>
    )
}
