import { FC, useContext } from 'react'

import { mdiDownload, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Button, H1, H3, Link, Text, Icon, BadgeVariantType } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../../app/tauriIcpUtils'
import { FooterWidget, SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'
import { AppSetupProgressBar } from '../components/AppSetupProgressBar'

import styles from './AppInstallExtensionsSetupStep.module.scss'

interface Extension {
    name: string
    status: ExtensionStatus
    iconURL: string
    docLink: string | null
    extensionDeepLink: string | null
}

enum ExtensionStatus {
    Beta = 'Beta',
    ComingSoon = 'Coming soon',
    Unknown = 'Unknown',
}

const EXTENSIONS: Extension[] = [
    {
        name: 'Visual Studio Code',
        status: ExtensionStatus.Beta,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/vscode-icon.png',
        docLink: null,
        extensionDeepLink: 'vscode:extension/sourcegraph.cody-ai',
    },
    {
        name: 'IntelliJ Idea',
        status: ExtensionStatus.ComingSoon,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/idea-icon.png',
        docLink: null,
        extensionDeepLink: null,
    },
    {
        name: 'NeoVim',
        status: ExtensionStatus.ComingSoon,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/neovim-icon.png',
        docLink: null,
        extensionDeepLink: null,
    },
]

export const AppInstallExtensionsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    const handleInstallExtensionClick = (extension: Extension): void => {
        if (extension.extensionDeepLink) {
            tauriShellOpen(extension.extensionDeepLink)
        }
    }

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
                {EXTENSIONS.map(extension => (
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
                                onClick={() => handleInstallExtensionClick(extension)}
                            >
                                <Icon svgPath={mdiDownload} aria-hidden={true} /> Install
                            </Button>
                        )}

                        {extension.docLink && (
                            <Link
                                to={extension.docLink}
                                target="_blank"
                                rel="noopener"
                                className={styles.extensionsActionLink}
                            >
                                <Icon svgPath={mdiOpenInNew} aria-hidden={true} /> Repo
                            </Link>
                        )}
                    </li>
                ))}

                <li className={styles.extensionsSuggestionLink}>
                    <Link
                        to="https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&title=Cody%20extension%20suggestion"
                        target="_blank"
                        rel="noopener"
                    >
                        Suggest our next extension
                    </Link>
                </li>
            </ul>

            <FooterWidget>
                <AppSetupProgressBar />
            </FooterWidget>
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
