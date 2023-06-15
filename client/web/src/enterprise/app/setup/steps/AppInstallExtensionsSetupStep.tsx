import { FC, useContext, useEffect, useState } from 'react'

import { mdiDownload, mdiOpenInNew } from '@mdi/js'
import { invoke } from '@tauri-apps/api/tauri'
import classNames from 'classnames'

import { Badge, Button, H1, H3, Link, Text, Icon, BadgeVariantType, LoadingSpinner } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../../app/tauriIcpUtils'
import { SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'

import styles from './AppInstallExtensionsSetupStep.module.scss'

interface Extension {
    name: string
    status: string
    iconURL: string
    docLink: string | null
    extensionDeepLink: string | null
}

enum ExtensionStatus {
    Beta = 'Beta',
    ComingSoon = 'Coming soon',
    Unknown = 'Unknown',
}

export const AppInstallExtensionsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)
    const [extensions, setExtensions] = useState<Extension[] | null>(null)

    useEffect(() => {
        fetchExtensionsConfiguration().then(extensions => setExtensions(extensions))
    }, [])

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Meet the extensions</H1>
                <Text className={styles.descriptionText}>
                    Ask Cody questions right within your editor. The Cody extension also has a fixup code feature,
                    recipes, and experimental completions.
                </Text>

                <Button size="lg" variant="primary" className={styles.descriptionNext} onClick={() => onNextStep()}>
                    Next →
                </Button>
            </div>

            <ul className={styles.extensions}>
                {extensions === null && (
                    <li className={styles.loading}>
                        <LoadingSpinner /> Loading extensions
                    </li>
                )}

                {extensions?.map(extension => (
                    <li key={extension.name} className={styles.extensionsItem}>
                        <img src={extension.iconURL} alt="" className={styles.extensionsIcon} />
                        <div className={styles.extensionsName}>
                            <H3 className="m-0">{extension.name}</H3>
                            <Badge variant={getBadgeStatus(extension.status)} small={true}>
                                {extension.status}
                            </Badge>
                        </div>

                        {extension.extensionDeepLink && (
                            <Button
                                variant="secondary"
                                outline={true}
                                size="sm"
                                onClick={() => tauriShellOpen(extension.extensionDeepLink)}
                            >
                                <Icon svgPath={mdiDownload} aria-hidden={true} /> Install
                            </Button>
                        )}

                        {extension.docLink && (
                            <Link to={extension.docLink} target="_blank" className={styles.extensionsActionLink}>
                                <Icon svgPath={mdiOpenInNew} aria-hidden={true} /> Repo
                            </Link>
                        )}
                    </li>
                ))}

                {extensions && (
                    <li className={styles.extensionsSuggestionLink}>
                        <Link
                            to="https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&title=Cody%20extension%20suggestion"
                            target="_blank"
                        >
                            Suggest our next extension
                        </Link>
                    </li>
                )}
            </ul>
        </div>
    )
}

function getBadgeStatus(status: ExtensionStatus): BadgeVariantType {
    switch (status) {
        case ExtensionStatus.Beta:
            return 'secondary'
        case ExtensionStatus.ComingSoon:
            return 'outlineSecondary'
        default:
            return 'outlineSecondary'
    }
}

let extensionCache: Extension[] | null = null

function fetchExtensionsConfiguration(): Promise<Extension[]> {
    if (extensionCache) {
        Promise.resolve(extensionCache)
    }

    return invoke<string>('get_extension_configuration').then(res => {
        const data = JSON.parse(res)
        extensionCache = data
        return data as Extension[]
    })
}
