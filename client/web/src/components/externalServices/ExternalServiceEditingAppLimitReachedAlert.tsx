import * as React from 'react'

import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Alert, H4, Text, Link } from '@sourcegraph/wildcard'

export const ExternalServiceEditingAppLimitReachedAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert variant="info">
        <H4>Code host limit reached</H4>
        <Text className="mb-0">
            Sourcegraph App is limited to one remote code host and up to 10 remote repositories. For more,{' '}
            <Link to={addSourcegraphAppOutboundUrlParameters('https://about.sourcegraph.com')}>
                get Sourcegraph Enterprise
            </Link>
            .
        </Text>
    </Alert>
)
