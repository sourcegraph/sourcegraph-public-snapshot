import React from 'react'

import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

export const ExcludeButton: React.FunctionComponent<{ handleExclude: () => void }> = ({ handleExclude }) => (
    <Button className="p-0 my-0 mx-2" data-tooltip="Omit this repository from batch spec file" onClick={handleExclude}>
        <Icon as={CloseIcon} />
    </Button>
)
