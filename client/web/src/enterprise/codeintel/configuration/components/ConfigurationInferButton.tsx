import React from 'react'

import { Button, Tooltip } from '@sourcegraph/wildcard'

export interface ConfigurationInferButtonProps {
    onClick?: () => void
}

export const ConfigurationInferButton: React.FunctionComponent<ConfigurationInferButtonProps> = ({ onClick }) => (
    <Tooltip content="Infer index configuration from HEAD">
        <Button type="button" variant="secondary" outline={true} className="ml-2" onClick={onClick}>
            Infer configuration
        </Button>
    </Tooltip>
)
