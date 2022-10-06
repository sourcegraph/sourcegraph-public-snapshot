import * as React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Tooltip } from '@sourcegraph/wildcard'

import { RegistryExtensionFields } from '../../../graphql-operations'

export const RegistryExtensionSourceBadge: React.FunctionComponent<
    React.PropsWithChildren<{
        extension: Pick<RegistryExtensionFields, 'remoteURL' | 'registryName' | 'isLocal'>
        showText?: boolean
        showRegistryName?: boolean
        className?: string
    }>
> = ({ extension, showText, showRegistryName, className = '' }) => (
    <Tooltip
        content={
            extension.isLocal
                ? 'Published on this site'
                : `Published on external extension registry ${extension.registryName}`
        }
    >
        <LinkOrSpan
            to={extension.remoteURL}
            target="_blank"
            rel="noopener noreferrer"
            className={classNames('text-muted text-nowrap d-inline-flex align-items-center', className)}
        >
            {showText && (extension.isLocal ? 'Local' : showRegistryName ? extension.registryName : 'External')}
        </LinkOrSpan>
    </Tooltip>
)
