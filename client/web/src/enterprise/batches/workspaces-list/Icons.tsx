import React from 'react'

import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'

import { Icon } from '@sourcegraph/wildcard'

export const CachedIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Icon
        role="img"
        data-tooltip="A cached result was found for this workspace."
        aria-label="A cached result was found for this workspace."
        as={ContentSaveIcon}
    />
)

export const ExcludeIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Icon
        role="img"
        data-tooltip="Your batch spec was modified to exclude this workspace. Preview again to update."
        aria-label="Your batch spec was modified to exclude this workspace. Preview again to update."
        as={DeleteIcon}
    />
)
