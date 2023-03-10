import * as React from 'react'

import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Alert, H4, Text, Link } from '@sourcegraph/wildcard'

export const ExternalServiceEditingAppLimitInPlaceAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert variant="info">
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
