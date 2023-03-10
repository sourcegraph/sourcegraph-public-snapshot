import { FC } from 'react'

import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Alert, H4, Text, Link } from '@sourcegraph/wildcard'

export const ExternalServiceEditingAppLimitInPlaceAlert: FC<{ className?: string }> = props => (
    <Alert variant="info" className={props.className}>
        <H4>Sourcegraph App limitations in place</H4>
        <Text className="mb-0">
            Only the first 10 remote repositories will be synchronized. For more,{' '}
            <Link to={addSourcegraphAppOutboundUrlParameters('https://about.sourcegraph.com')}>
                get Sourcegraph Enterprise
            </Link>
            .
        </Text>
    </Alert>
)
