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

/** Default icon for Sourcegraph extensions */
const defaultIcon = (
    <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
        <mask id="mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="48" height="48">
            <path
                d="M41.407 23.2093C39.7671 23.6779 37.7256 26.4305 37.6503 25.301C37.6169 24.7405 37.4746 23.4436 37.4663 22.1217C37.4412 18.3483 37.6587 12.6923 37.6587 12.6923C37.6587 12.4999 37.642 12.3158 37.6001 12.1401C37.4997 11.5796 37.232 11.1361 36.8806 10.7847C36.8555 10.7596 36.8304 10.7429 36.8053 10.7178C36.7133 10.6341 36.6129 10.5505 36.5125 10.4752C36.4037 10.3999 36.2782 10.3329 36.1611 10.266C36.1109 10.2409 36.069 10.2158 36.0188 10.1907C35.9937 10.1823 35.9686 10.174 35.9519 10.1656C35.4332 9.93969 34.8558 9.82256 34.3287 9.82256L32.7809 9.85602C32.7809 9.85602 30.8147 9.92296 28.5557 9.97316H22.8914C22.4229 9.94806 22.1133 9.91459 21.9962 9.85602C20.7412 9.27035 23.4771 7.21213 23.887 6.02405C24.958 2.89489 22.1468 0 18.8335 0C15.5119 0 12.7007 2.89489 13.7884 6.04079C14.2067 7.22887 17.0765 9.32055 15.6793 9.87276C15.4952 9.94806 14.4494 9.98152 13.0856 9.98152H11.7971C8.69305 9.97316 4.87781 9.88112 4.87781 9.88112L3.32996 9.84766C1.94108 9.84766 0.326303 10.5003 0.0836674 12.0648C0.033467 12.274 0 12.7007 0 12.7007L0.0251002 24.3472C0.0418337 24.5397 0.0585672 24.6819 0.0836674 24.7321C0.635872 26.1293 3.68137 22.4814 4.86944 22.0631C8.00697 20.9671 9.90622 25.6859 9.90622 29.0159C9.90622 32.3542 7.0197 35.1654 3.88217 34.0777C2.69409 33.6594 0.644239 30.9235 0.0585672 32.1785C0.033467 32.2203 0.0167335 32.304 0 32.4211V45.331C0.0251002 46.8119 1.22154 48 2.69409 48H15.194C15.3613 47.9833 15.4868 47.9665 15.537 47.9498C16.9343 47.3892 15.0183 46.084 14.6418 44.7537C13.7298 41.5492 15.5036 40.9887 18.8335 40.9887C22.1635 40.9887 23.7364 41.8253 23.7364 44.1095C23.7364 45.2808 20.5906 47.3558 21.8456 47.9414C21.8874 47.9582 21.9627 47.9749 22.0631 47.9916H34.9395C36.4372 47.9916 37.642 46.7785 37.642 45.2808C37.642 45.2808 37.3742 33.2243 37.5918 31.71C37.7507 30.6139 40.2942 32.8897 41.407 33.3582C45.5485 35.0817 47.4813 31.6263 47.4813 28.2712C47.4813 24.9162 44.6282 22.289 41.407 23.2093Z"
                fill="#D5E4F6"
            />
        </mask>
        <g mask="url(#mask0)">
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M9.94394 33.0201L-3.26866 28.731C-6.11988 27.8033 -7.66723 24.784 -6.72289 21.9832C-5.78083 19.1824 -2.70661 17.6661 0.144605 18.5893L11.5793 22.3029L8.3598 10.7511C7.56565 7.90572 9.27228 4.96891 12.1667 4.19066C15.0635 3.41242 18.0535 5.08709 18.8476 7.93025L22.0894 19.5601L30.183 10.5904C32.1764 8.38278 35.6101 8.17986 37.8583 10.1355C40.1065 12.0911 40.3136 15.465 38.3225 17.6727L29.0305 27.9705L42.2689 32.2699C45.1178 33.1931 46.6674 36.2124 45.7231 39.0132C44.7788 41.8117 41.7046 43.3326 38.8533 42.4049L27.4229 38.6943L30.6439 50.2498C31.4358 53.093 29.7314 56.0298 26.8347 56.8103C23.938 57.5863 20.9479 55.9116 20.1538 53.0684L16.9043 41.4092L8.78639 50.4058C6.79531 52.6134 3.35929 52.8186 1.11108 50.8629C-1.13712 48.9073 -1.34419 45.5311 0.64688 43.3235L9.94394 33.0201Z"
                fill="url(#paint0_linear)"
            />
        </g>
        <path
            d="M41.407 23.2093C39.7671 23.6779 37.7256 26.4305 37.6503 25.301C37.6169 24.7405 37.4746 23.4436 37.4663 22.1217C37.4412 18.3483 37.6587 12.6923 37.6587 12.6923C37.6587 12.4999 37.642 12.3158 37.6001 12.1401C37.4997 11.5796 37.232 11.1361 36.8806 10.7847C36.8555 10.7596 36.8304 10.7429 36.8053 10.7178C36.7133 10.6341 36.6129 10.5505 36.5125 10.4752C36.4037 10.3999 36.2782 10.3329 36.1611 10.266C36.1109 10.2409 36.069 10.2158 36.0188 10.1907C35.9937 10.1823 35.9686 10.174 35.9519 10.1656C35.4332 9.93969 34.8558 9.82256 34.3287 9.82256L32.7809 9.85602C32.7809 9.85602 30.8147 9.92296 28.5557 9.97316H22.8914C22.4229 9.94806 22.1133 9.91459 21.9962 9.85602C20.7412 9.27035 23.4771 7.21213 23.887 6.02405C24.958 2.89489 22.1468 0 18.8335 0C15.5119 0 12.7007 2.89489 13.7884 6.04079C14.2067 7.22887 17.0765 9.32055 15.6793 9.87276C15.4952 9.94806 14.4494 9.98152 13.0856 9.98152H11.7971C8.69305 9.97316 4.87781 9.88112 4.87781 9.88112L3.32996 9.84766C1.94108 9.84766 0.326303 10.5003 0.0836674 12.0648C0.033467 12.274 0 12.7007 0 12.7007L0.0251002 24.3472C0.0418337 24.5397 0.0585672 24.6819 0.0836674 24.7321C0.635872 26.1293 3.68137 22.4814 4.86944 22.0631C8.00697 20.9671 9.90622 25.6859 9.90622 29.0159C9.90622 32.3542 7.0197 35.1654 3.88217 34.0777C2.69409 33.6594 0.644239 30.9235 0.0585672 32.1785C0.033467 32.2203 0.0167335 32.304 0 32.4211V45.331C0.0251002 46.8119 1.22154 48 2.69409 48H15.194C15.3613 47.9833 15.4868 47.9665 15.537 47.9498C16.9343 47.3892 15.0183 46.084 14.6418 44.7537C13.7298 41.5492 15.5036 40.9887 18.8335 40.9887C22.1635 40.9887 23.7364 41.8253 23.7364 44.1095C23.7364 45.2808 20.5906 47.3558 21.8456 47.9414C21.8874 47.9582 21.9627 47.9749 22.0631 47.9916H34.9395C36.4372 47.9916 37.642 46.7785 37.642 45.2808C37.642 45.2808 37.3742 33.2243 37.5918 31.71C37.7507 30.6139 40.2942 32.8897 41.407 33.3582C45.5485 35.0817 47.4813 31.6263 47.4813 28.2712C47.4813 24.9162 44.6282 22.289 41.407 23.2093Z"
            fill="#95A5C6"
            fillOpacity="0.24"
        />
        <defs>
            <linearGradient id="paint0_linear" x1="19.5" y1="4" x2="19.5" y2="57" gradientUnits="userSpaceOnUse">
                <stop stopColor="#A2B0CD" stopOpacity="0" />
                <stop offset="1" stopColor="#A2B0CD" stopOpacity="0.3" />
            </linearGradient>
        </defs>
    </svg>
)

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
                        {icon ? (
                            <img className="extension-card__icon mr-2" src={icon} />
                        ) : (
                            publisher === 'sourcegraph' && defaultIcon
                        )}
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
