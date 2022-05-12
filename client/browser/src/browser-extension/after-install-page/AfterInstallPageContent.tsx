import React, { VideoHTMLAttributes } from 'react'

import classNames from 'classnames'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import LockIcon from 'mdi-react/LockIcon'

import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Icon, Typography } from '@sourcegraph/wildcard'

import { getPlatformName } from '../../shared/util/context'

import styles from './AfterInstallPageContent.module.scss'

const Video: React.FunctionComponent<
    React.PropsWithChildren<
        { name: string } & Pick<VideoHTMLAttributes<HTMLVideoElement>, 'width' | 'height'> & ThemeProps
    >
> = ({ name, isLightTheme, width, height }) => {
    const suffix = isLightTheme ? 'Light' : 'Dark'
    return (
        <video
            className="w-100 h-auto cursor-pointer"
            width={width}
            height={height}
            autoPlay={true}
            loop={true}
            muted={true}
            playsInline={true}
            onClick={event => event.currentTarget.requestFullscreen()}
            // Add a key on the theme to force React to render a new <video> element when the theme changes
            key={name + suffix}
        >
            <source
                src={`https://storage.googleapis.com/sourcegraph-assets/code-host-integration/${name}${suffix}.webm`}
                type="video/webm"
            />
            <source
                src={`https://storage.googleapis.com/sourcegraph-assets/code-host-integration/${name}${suffix}.mp4`}
                type="video/mp4"
            />
        </video>
    )
}

export const AfterInstallPageContent: React.FunctionComponent<React.PropsWithChildren<ThemeProps>> = props => {
    // Safari does not support the search shortcut. So don't show the feature.
    const isSafari = getPlatformName() === 'safari-extension'
    const showSearchShortcut = !isSafari

    return (
        <div className="after-install-page-content" data-testid="after-install-page-content">
            <div className="d-flex w-100 p-3 justify-content-between align-items-center">
                <Link to="https://sourcegraph.com/search" target="_blank" rel="noopener">
                    <SourcegraphLogo className={styles.sourcegraphLogo} />
                </Link>
                <Link to="https://docs.sourcegraph.com/integration/browser_extension" target="_blank" rel="noopener">
                    Browser extension docs <Icon role="img" as={ExternalLinkIcon} aria-hidden={true} />
                </Link>
            </div>

            <div className="container mt-3">
                <Typography.H1>🎉 You’ve just installed the Sourcegraph browser extension!</Typography.H1>
                <p className="lead mb-0">We’ve gathered the most important information that will get your started:</p>
            </div>

            <section className="border-bottom py-5">
                <div className="container">
                    <Typography.H2 className="mb-4">How do I use the extension?</Typography.H2>
                    <div className="row">
                        <div className="col-md-6">
                            <Typography.H3>Code intelligence on your code host</Typography.H3>
                            <p>
                                Sourcegraph browser extension adds code intelligence to files and diffs on GitHub,
                                GitHub Enterprise, GitLab, Phabricator, Bitbucket Server, and Gerrit.
                            </p>
                            <Video {...props} name="CodeIntelligenceOnCodeHost" width={1760} height={1060} />
                        </div>
                        {showSearchShortcut && (
                            <div className="col-md-6 mt-4 mt-md-0">
                                <Typography.H3>Search shortcut in the URL location bar</Typography.H3>
                                <p>
                                    Type <code>src</code>
                                    <kbd>tab</kbd> in the address bar of your browser to search for queries on
                                    Sourcegraph.
                                </p>
                                <Video {...props} name="BrowserShortcut" width={1196} height={720} />
                            </div>
                        )}
                    </div>
                </div>
            </section>

            <section className="border-bottom py-5">
                <div className="container">
                    <div className="row">
                        <div className="col-md-6 d-flex flex-column">
                            <Typography.H2 className="mb-4">Make it work on your codehost</Typography.H2>
                            <div className="bg-2 rounded p-3 mb-3 d-flex flex-column justify-content-center">
                                <Typography.H3 className={classNames('mb-3', styles.codeHostTitles)}>
                                    <Icon
                                        role="img"
                                        className={styles.codeHostLogo}
                                        as={GithubIcon}
                                        aria-hidden={true}
                                    />{' '}
                                    github.com
                                </Typography.H3>
                                <p className="m-0">
                                    <Icon role="img" as={CheckIcon} aria-hidden={true} /> No action required.Your
                                    extension works here by default.
                                </p>
                            </div>
                            <div className="bg-2 rounded p-3 d-flex flex-column justify-content-center">
                                <Typography.H3 className={classNames('d-flex flex-wrap', styles.codeHostTitles)}>
                                    <div className="mr-5 mb-3">
                                        <Icon
                                            role="img"
                                            className={styles.codeHostLogo}
                                            as={GithubIcon}
                                            aria-hidden={true}
                                        />{' '}
                                        GitHub Enterprise
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <Icon
                                            role="img"
                                            className={styles.codeHostLogo}
                                            as={GitlabIcon}
                                            aria-hidden={true}
                                        />{' '}
                                        GitLab
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <Icon
                                            role="img"
                                            className={styles.codeHostLogo}
                                            as={BitbucketIcon}
                                            aria-hidden={true}
                                        />{' '}
                                        Bitbucket Server
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <Icon
                                            role="img"
                                            className={styles.codeHostLogo}
                                            as={PhabricatorIcon}
                                            aria-hidden={true}
                                        />{' '}
                                        Phabricator
                                    </div>
                                </Typography.H3>
                                <p>Your extension needs explicit permissions to your code host:</p>
                                <ol className="m-0">
                                    <li>Navigate to any page on your code host.</li>
                                    <li>
                                        Click the{' '}
                                        <q>
                                            <strong>Grant permissions</strong>
                                        </q>{' '}
                                        button.
                                    </li>
                                    <li>
                                        Click{' '}
                                        <q>
                                            <strong>Allow</strong>
                                        </q>{' '}
                                        in the permissions request popup.
                                    </li>
                                </ol>
                            </div>
                        </div>
                        <div className="col-md-6 mt-4 mt-md-0">
                            <Video {...props} name="GrantPermissions" width={1762} height={1384} />
                        </div>
                    </div>
                </div>
            </section>

            <section className="border-bottom py-5">
                <div className="container">
                    <div className="row">
                        <div className="col-md-6 d-flex flex-column">
                            <Typography.H2 className="mb-4">Make it work for private code</Typography.H2>
                            <p>By default, the browser extension works only for public code.</p>
                            <div className="d-flex align-items-center">
                                <div className="bg-3 rounded-circle p-2">
                                    <Icon role="img" as={LockIcon} aria-hidden={true} />
                                </div>
                                <p className="m-0 ml-3">
                                    To use the browser extension with your private repositories, you need to set up a{' '}
                                    <strong>private Sourcegraph instance</strong> and connect the extension to it.
                                </p>
                            </div>
                            <div className="bg-2 rounded p-3 mt-4 d-flex flex-column justify-content-around">
                                <p>Follow these instructions:</p>
                                <ol className="m-0 d-flex flex-column justify-content-around">
                                    <li>
                                        <strong>Install Sourcegraph</strong> (
                                        <Link
                                            to="https://docs.sourcegraph.com/admin/install"
                                            target="_blank"
                                            rel="noopener"
                                        >
                                            visit our docs for instructions
                                        </Link>
                                        ).Skip this step if you already have a private Sourcegraph instance.
                                    </li>
                                    <li>
                                        Click the Sourcegraph extension icon in the browser toolbar to{' '}
                                        <Link to="./options.html" rel="noopener" target="_blank">
                                            open the settings page
                                        </Link>
                                        .
                                    </li>
                                    <li>
                                        Enter the <strong>URL</strong> (including the protocol) of your Sourcegraph
                                        instance (such as <q>https://sourcegraph.example.com</q>).
                                    </li>
                                    <li>Make sure a green checkmark appears in the input field.</li>
                                </ol>
                            </div>
                        </div>
                        <div className="col-md-6 mt-4 mt-md-0">
                            <Video {...props} name="PrivateInstance" width={1764} height={1390} />
                        </div>
                    </div>
                </div>
            </section>

            <section className="py-5">
                <div className="container">
                    <Typography.H2 className="mb-4">Additional resources</Typography.H2>
                    <div className="d-flex w-100 align-items-center">
                        <div className="bg-3 rounded-circle p-2">
                            <Icon role="img" as={BookOpenPageVariantIcon} aria-hidden={true} />
                        </div>
                        <p className="m-0 ml-3">
                            Read the{' '}
                            <Link
                                to="https://docs.sourcegraph.com/integration/browser_extension"
                                rel="noopener"
                                target="_blank"
                            >
                                Sourcegraph docs
                            </Link>{' '}
                            to learn more about how we respect your privacy, troubleshooting and extension features.
                        </p>
                    </div>
                </div>
            </section>
        </div>
    )
}
