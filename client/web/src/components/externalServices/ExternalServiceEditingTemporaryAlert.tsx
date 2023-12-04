import type { FC } from 'react'

import { Alert, H4, Code, Text } from '@sourcegraph/wildcard'

export const ExternalServiceEditingTemporaryAlert: FC<{ className?: string }> = props => (
    <Alert variant="warning" className={props.className}>
        <H4>Edits will be reset when restarting</H4>
        <Text className="mb-0">
            Environment variable <Code>EXTSVC_CONFIG_ALLOW_EDITS</Code> is set along with{' '}
            <Code>EXTSVC_CONFIG_FILE</Code>. Every change made through the UI will be undone with the next restart of
            the Sourcegraph instance and the code host connection configuration is reset to the contents of{' '}
            <Code>EXTSVC_CONFIG_FILE</Code>.
        </Text>
    </Alert>
)
