import * as React from 'react'

import { Alert, H4, Code, Text } from '@sourcegraph/wildcard'

export const ExternalServiceEditingTemporaryAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert variant="warning">
        <H4>Edits will be reset when restarting</H4>
        <Text className="mb-0">
            Environment variable <Code>EXTSVC_CONFIG_ALLOW_EDITS</Code> is set along with{' '}
            <Code>EXTSVC_CONFIG_FILE</Code>. Every change made through the UI will be undone with the next restart of
            the Sourcegraph instance and the code host connection configuration is reset to the contents of{' '}
            <Code>EXTSVC_CONFIG_FILE</Code>.
        </Text>
    </Alert>
)
