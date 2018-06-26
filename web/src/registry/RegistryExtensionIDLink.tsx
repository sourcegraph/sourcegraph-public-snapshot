import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { RegistryExtensionNodeDisplayProps } from './RegistryExtensionsPage'

interface Props extends Pick<RegistryExtensionNodeDisplayProps, 'showExtensionID'> {
    extension: Pick<
        GQL.IRegistryExtension,
        'url' | 'registryName' | 'extensionID' | 'extensionIDWithoutRegistry' | 'name'
    >

    showRegistryMuted?: boolean
    className?: string
}

/** Displays a link to the extension. */
export const RegistryExtensionIDLink: React.SFC<Props> = ({
    extension,
    showExtensionID = 'extensionID',
    showRegistryMuted,
    className = '',
}) => {
    if (showRegistryMuted && showExtensionID === 'extensionID') {
        showExtensionID = 'extensionIDWithoutRegistry'
    }
    return (
        <Link to={extension.url} className={`registry-extension-id-link ${className}`} title={extension.extensionID}>
            {showRegistryMuted &&
                showExtensionID === 'extensionIDWithoutRegistry' &&
                extension.registryName && (
                    <span className="text-muted font-weight-normal registry-extension-id-link__span">
                        {extension.registryName}/
                    </span>
                )}
            <span className="registry-extension-id-link__span">{extension[showExtensionID]}</span>
        </Link>
    )
}
