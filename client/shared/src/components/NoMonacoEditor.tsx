import React from 'react'

import { Code, Text } from '@sourcegraph/wildcard'

/**
 * Used when the env var `DEV_WEB_BUILDER_OMIT_SLOW_DEPS` is set.
 */
export const MonacoEditor: React.FunctionComponent = () => (
    <Text className="border border-danger p-2">
        Monaco editor is not included in this bundle because the environment variable{' '}
        <Code>DEV_WEB_BUILDER_OMIT_SLOW_DEPS</Code> is set.
    </Text>
)
