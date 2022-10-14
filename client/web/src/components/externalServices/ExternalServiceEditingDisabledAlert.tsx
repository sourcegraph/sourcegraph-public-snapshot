import * as React from 'react'

import { Alert, H4, Code, Text } from '@sourcegraph/wildcard'

export const ExternalServiceEditingDisabledAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert variant="info">
        <H4>Editing through UI disabled</H4>
        <Text className="mb-0">
            Environment variable <Code>EXTSVC_CONFIG_FILE</Code> is set. You can't create or edit code host connections
            when <Code>EXTSVC_CONFIG_FILE</Code> is set. If you also set <Code>EXTSVC_CONFIG_ALLOW_EDITS</Code> to{' '}
            <Code>"true"</Code> you can edit code host connections, but changes will be discarded with the next restart
            of the Sourcegraph instance.
        </Text>
    </Alert>
)
