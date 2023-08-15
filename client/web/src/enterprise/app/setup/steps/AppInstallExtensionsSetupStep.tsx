import { type FC, useContext } from 'react'

import { mdiDownload, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Button, H1, H3, Link, Text, Icon } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../../app/tauriIcpUtils'
import { FooterWidget, SetupStepsContext, type StepComponentProps } from '../../../../setup-wizard/components'
import { AppSetupProgressBar } from '../components/AppSetupProgressBar'

import styles from './AppInstallExtensionsSetupStep.module.scss'

interface Extension {
    name: string
    status: ExtensionStatus
    iconURL: string
    docLink?: string
    installHref?: string
    installLabel?: string
}

enum ExtensionStatus {
    Beta = 'Beta',
    ComingSoon = 'Coming soon',
    Unknown = 'Unknown',
    Experimental = 'Experimental',
    GA = '',
}

const EXTENSIONS: Extension[] = [
    {
        name: 'Visual Studio Code',
        status: ExtensionStatus.GA,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/vscode-icon.png',
        installHref: 'vscode:extension/sourcegraph.cody-ai',
        installLabel: 'Install extension',
    },
    {
        name: 'IntelliJ Idea',
        status: ExtensionStatus.Experimental,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/idea-icon.png',
        installHref: 'https://plugins.jetbrains.com/plugin/9682-sourcegraph',
        installLabel: 'Install plugin',
    },
    {
        name: 'NeoVim',
        status: ExtensionStatus.ComingSoon,
        iconURL: 'https://storage.googleapis.com/sourcegraph-assets/setup/neovim-icon.png',
    },
]

export const AppInstallExtensionsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    const handleInstallExtensionClick = (extension: Extension): void => {
        if (extension.installHref) {
            tauriShellOpen(extension.installHref)
        }
    }

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Install an extension</H1>
                <Text className={styles.descriptionText}>
                    Use Cody from within your editor, using one of the Cody extensions.
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
                            <Badge variant="outlineSecondary" small={true}>
                                {extension.status}
                            </Badge>
                        </div>

                        {extension.installHref && (
                            <Button
                                variant="secondary"
                                outline={true}
                                size="sm"
                                onClick={() => handleInstallExtensionClick(extension)}
                            >
                                <Icon svgPath={mdiDownload} aria-hidden={true} /> {extension.installLabel}
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
                    <span>
                        Your editor not listed here?
                        <br />
                        <Link
                            to="https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&title=Cody%20extension%20suggestion"
                            target="_blank"
                            rel="noopener"
                        >
                            Suggest an extension
                        </Link>
                    </span>
                </li>
            </ul>

            <FooterWidget>
                <AppSetupProgressBar />
            </FooterWidget>
        </div>
    )
}
