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
import { DefaultIconEnabled, DefaultIcon } from './icons'
import { ThemeProps } from '../../../shared/src/theme'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'>, ThemeProps {
    node: Pick<
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

/** ms after which to remove visual feedback */
const FEEDBACK_DELAY = 50000

/** Displays an extension as a card. */
export const ExtensionCard = React.memo<Props>(function ExtensionCard({
    node: extension,
    settingsCascade,
    platformContext,
    subject,
    enabled,
    isLightTheme,
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

    /**
     * When extension enablement state changes, display visual feedback for $delay seconds.
     * Clear the timeout when the component unmounts or the extension is toggled again.
     */
    const [change, setChange] = React.useState<'enabled' | 'disabled' | null>(null)
    const timeoutReference = React.useRef<number | undefined>()

    React.useEffect(() => () => clearTimeout(timeoutReference.current), [])

    const onToggleChange = React.useCallback((enabled: boolean): void => {
        if (timeoutReference.current) {
            clearTimeout(timeoutReference.current)
        }
        setChange(enabled ? 'enabled' : 'disabled')
        timeoutReference.current = window.setTimeout(() => setChange(null), FEEDBACK_DELAY)
    }, [])

    return (
        <div className="d-flex">
            <div className={`extension-card card ${change === 'enabled' ? 'alert alert-success p-0 m-0' : ''}`}>
                <div
                    className="card-body extension-card__body d-flex position-relative"
                    // Prevent toggle clicks from propagating to the stretched-link (and
                    // navigating to the extension detail page).
                    onClick={stopPropagation}
                >
                    {/* Item 1: Icon */}
                    <div className="flex-shrink-0 mr-2">
                        {icon ? (
                            <img className="extension-card__icon" src={icon} />
                        ) : publisher === 'sourcegraph' ? (
                            change === 'enabled' ? (
                                <DefaultIconEnabled isLightTheme={isLightTheme} />
                            ) : (
                                <DefaultIcon />
                            )
                        ) : null}
                    </div>
                    {/* Item 2: Text */}
                    {change === 'enabled' ? (
                        <span className="extension-card__enabled-feedback mr-1">
                            <span className="alert-success font-weight-bold">{name}</span> is now enabled in code search
                            results.{' '}
                            <Link to={`/extensions/${extension.id}`} className="extension-card__link alert-link">
                                See how it works
                            </Link>
                        </span>
                    ) : (
                        <div className="text-truncate w-100">
                            <div className="d-flex align-items-center">
                                <span className="mb-0 mr-1 text-truncate flex-1">
                                    <Link
                                        to={`/extensions/${
                                            extension.registryExtension
                                                ? extension.registryExtension.extensionIDWithoutRegistry
                                                : extension.id
                                        }`}
                                    >
                                        <strong>{name}</strong>
                                    </Link>
                                    <span className="text-muted"> by {publisher}</span>
                                </span>

                                {extension.registryExtension?.isWorkInProgress && (
                                    <WorkInProgressBadge
                                        viewerCanAdminister={extension.registryExtension.viewerCanAdminister}
                                    />
                                )}
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
                    )}
                    {/* Item 3: Toggle */}
                    {subject &&
                        (subject.viewerCanAdminister ? (
                            <ExtensionToggle
                                extensionID={extension.id}
                                enabled={enabled}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                className="extension-card__toggle flex-shrink-0 align-self-start"
                                onToggleChange={onToggleChange}
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
                {/* Visual feedback: alert when extension is disabled */}
                {change === 'disabled' && (
                    <div className="alert alert-secondary px-2 py-1 extension-card__disabled-feedback">
                        <strong>{name}</strong> is off
                    </div>
                )}
            </div>
        </div>
    )
},
areEqual)

/** Custom compareFunction for ExtensionCard */
function areEqual(oldProps: Props, newProps: Props): boolean {
    return oldProps.enabled === newProps.enabled
}
