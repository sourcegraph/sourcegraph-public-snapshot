import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export interface UnsupportedProps {}

export const Unsupported: React.FunctionComponent<React.PropsWithChildren<UnsupportedProps>> = () => (
    <div className="px-2 py-1">
        <div className="d-flex align-items-center">
            <div className="px-2 py-1 text-uppercase">
                <Badge variant="outlineSecondary">Unsupported</Badge>
            </div>
            <div className="px-2 py-1">
                <p className="mb-0">No language detected</p>
            </div>
        </div>
    </div>
)
