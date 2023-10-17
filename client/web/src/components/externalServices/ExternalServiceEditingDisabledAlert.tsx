import type { FC } from 'react'

import { Alert, H4, Code, Text, Link } from '@sourcegraph/wildcard'

export const ExternalServiceEditingDisabledAlert: FC<{ className?: string }> = props => (
    <Alert variant="info" className={props.className}>
        <H4>Editing through UI disabled</H4>
        <Text className="mb-0">
            Environment variable <Code>EXTSVC_CONFIG_FILE</Code> is set.{' '}
            <Link to="/help/admin/config/advanced_config_file#code-host-configuration">
                You can't create or edit code host connections when <Code>EXTSVC_CONFIG_FILE</Code> is set.
            </Link>{' '}
            If you also set <Code>EXTSVC_CONFIG_ALLOW_EDITS</Code> to <Code>"true"</Code> you can edit code host
            connections, but changes will be discarded with the next restart of the Sourcegraph instance.
        </Text>
    </Alert>
)
