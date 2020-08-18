import WarningIcon from 'mdi-react/WarningIcon'
import * as React from 'react'
import { ConfiguredRegistryExtension } from '../../../shared/src/extensions/extension'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { ExtensionManifest } from '../../../shared/src/schema/extensionSchema'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { isExtensionAdded } from './extension/extension'
import { ExtensionConfigurationState } from './extension/ExtensionConfigurationState'
import { WorkInProgressBadge } from './extension/WorkInProgressBadge'
import { ExtensionToggle } from './ExtensionToggle'
import { isEncodedImage } from '../../../shared/src/util/icon'
import { Link } from 'react-router-dom'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    extension: Pick<
        ConfiguredRegistryExtension<
            Pick<
                GQL.IRegistryExtension,
                'id' | 'extensionIDWithoutRegistry' | 'isWorkInProgress' | 'viewerCanAdminister' | 'url'
            >
        >,
        'id' | 'manifest' | 'registryExtension'
    >
    subject: Pick<GQL.SettingsSubject, 'id' | 'viewerCanAdminister'>
    enabled: boolean
}

const stopPropagation: React.MouseEventHandler<HTMLElement> = event => {
    event.stopPropagation()
}

/** Displays an extension as a card. */
export const ExtensionCard = React.memo<Props>(function ExtensionCard({
    extension,
    settingsCascade,
    platformContext,
    subject,
    enabled,
}) {
    const manifest: ExtensionManifest | undefined =
        extension.manifest && !isErrorLike(extension.manifest) ? extension.manifest : undefined

    const icon = React.useMemo(() => {
        let url: string | undefined
        if (manifest?.icon && isEncodedImage(manifest.icon)) {
            url = manifest.icon
        }
        return url
    }, [manifest])

    const [publisher, name] = React.useMemo(() => {
        const id = extension.registryExtension ? extension.registryExtension.extensionIDWithoutRegistry : extension.id

        return id.split('/')
    }, [extension])

    return (
        <div className="d-flex">
            <div className="extension-card card">
                <div
                    className="card-body extension-card__body d-flex flex-column position-relative"
                    // Prevent toggle clicks from propagating to the stretched-link (and
                    // navigating to the extension detail page).
                    onClick={stopPropagation}
                >
                    <div className="d-flex">
                        {icon && <img className="extension-card__icon mr-2" src={icon} />}
                        <div className="text-truncate w-100">
                            <div className="d-flex align-items-center">
                                <h4 className="card-title extension-card__body-title mb-0 mr-1 text-truncate font-weight-normal flex-1">
                                    <Link
                                        to={`/extensions/${
                                            extension.registryExtension
                                                ? extension.registryExtension.extensionIDWithoutRegistry
                                                : extension.id
                                        }`}
                                        className="extension-card__name"
                                    >
                                        {name}
                                    </Link>
                                    <span className="extension-card__publisher"> by {publisher}</span>
                                </h4>
                                {extension.registryExtension?.isWorkInProgress && (
                                    <WorkInProgressBadge
                                        viewerCanAdminister={extension.registryExtension.viewerCanAdminister}
                                    />
                                )}
                                {subject &&
                                    (subject.viewerCanAdminister ? (
                                        <ExtensionToggle
                                            extensionID={extension.id}
                                            enabled={enabled}
                                            settingsCascade={settingsCascade}
                                            platformContext={platformContext}
                                            className="extension-card__toggle"
                                        />
                                    ) : (
                                        <ExtensionConfigurationState
                                            isAdded={isExtensionAdded(settingsCascade.final, extension.id)}
                                            isEnabled={enabled}
                                            enabledIconOnly={true}
                                            className="small"
                                        />
                                    ))}
                            </div>
                            <div className="mt-1">
                                {extension.manifest ? (
                                    isErrorLike(extension.manifest) ? (
                                        <span className="text-danger small" title={extension.manifest.message}>
                                            <WarningIcon className="icon-inline" /> Invalid manifest
                                        </span>
                                    ) : (
                                        extension.manifest.description && (
                                            <div className="text-muted text-truncate">
                                                {extension.manifest.description}
                                            </div>
                                        )
                                    )
                                ) : (
                                    <span className="text-warning small">
                                        <WarningIcon className="icon-inline" /> No manifest
                                    </span>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
},
areEqual)

/** Custom compareFunction for ExtensionCard */
function areEqual(oldProps: Props, newProps: Props): boolean {
    return oldProps.enabled === newProps.enabled
}
