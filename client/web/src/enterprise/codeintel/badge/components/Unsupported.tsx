import React from 'react'

import { Badge, Text } from '@sourcegraph/wildcard'

export interface UnsupportedProps {}

export const Unsupported: React.FunctionComponent<React.PropsWithChildren<UnsupportedProps>> = () => (
    <div className="px-2 py-1">
        <div className="d-flex align-items-center">
            <div className="px-2 py-1 text-uppercase">
                <Badge variant="outlineSecondary">Unsupported</Badge>
            </div>
            <div className="px-2 py-1">
                <Text className="mb-0">No language detected</Text>
            </div>
        </div>
    </div>
)
