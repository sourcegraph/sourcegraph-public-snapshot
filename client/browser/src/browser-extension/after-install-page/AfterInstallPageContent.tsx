import BitbucketIcon from 'mdi-react/BitbucketIcon'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import LockIcon from 'mdi-react/LockIcon'
import React, { VideoHTMLAttributes } from 'react'

import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { getPlatformName } from '../../shared/util/context'

const Video: React.FunctionComponent<
    { name: string } & Pick<VideoHTMLAttributes<HTMLVideoElement>, 'width' | 'height'> & ThemeProps
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

export const AfterInstallPageContent: React.FunctionComponent<ThemeProps> = props => {
    // Safari does not support the search shortcut. So don't show the feature.
    const isSafari = getPlatformName() === 'safari-extension'
    const showSearchShortcut = !isSafari

    return (
        <div className="after-install-page-content">
            <div className="d-flex w-100 p-3 justify-content-between align-items-center">
                <a href="https://sourcegraph.com/search" target="_blank" rel="noopener">
                    <SourcegraphLogo className="after-install-page-content__sourcegraph-logo" />
                </a>
                <a href="https://docs.sourcegraph.com/integration/browser_extension" target="_blank" rel="noopener">
                    Browser extension docs <ExternalLinkIcon className="icon-inline" />
                </a>
            </div>

            <div className="container mt-3">
                <h1>🎉 You’ve just installed the Sourcegraph browser extension!</h1>
                <p className="lead mb-0">We’ve gathered the most important information that will get your started:</p>
            </div>

            <section className="border-bottom py-5">
                <div className="container">
                    <h2 className="mb-4">How do I use the extension?</h2>
                    <div className="row">
                        <div className="col-md-6">
                            <h3>Code intelligence on your code host</h3>
                            <p>
                                Sourcegraph browser extension adds code intelligence to files and diffs on GitHub,
                                GitHub Enterprise, GitLab, Phabricator, Bitbucket Server, and Gerrit.
                            </p>
                            <Video {...props} name="CodeIntelligenceOnCodeHost" width={1760} height={1060} />
                        </div>
                        {showSearchShortcut && (
                            <div className="col-md-6 mt-4 mt-md-0">
                                <h3>Search shortcut in the URL location bar</h3>
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
                            <h2 className="mb-4">Make it work on your codehost</h2>
                            <div className="bg-2 rounded p-3 mb-3 d-flex flex-column justify-content-center">
                                <h3 className="mb-3 after-install-page-content__code-host-titles">
                                    <GithubIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                    github.com
                                </h3>
                                <p className="m-0">
                                    <CheckIcon className="icon-inline" /> No action required. Your extension works here
                                    by default.
                                </p>
                            </div>
                            <div className="bg-2 rounded p-3 d-flex flex-column justify-content-center">
                                <h3 className="d-flex flex-wrap after-install-page-content__code-host-titles">
                                    <div className="mr-5 mb-3">
                                        <GithubIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                        GitHub Enterprise
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <GitlabIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                        GitLab
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <BitbucketIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                        Bitbucket Server
                                    </div>
                                    <div className="mr-5 mb-3">
                                        <PhabricatorIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                        Phabricator
                                    </div>
                                </h3>
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
                    <h2 className="mb-4">Make it work for private code</h2>
                    <div className="row">
                        <div className="col-md-6 d-flex flex-column">
                            <div className="bg-2 rounded p-3 mb-3 d-flex flex-column justify-content-center">
                                <h3 className="mb-3 after-install-page-content__code-host-titles">
                                    Private code synced with Sourcegraph cloud
                                </h3>
                                <p className="m-0">
                                    <CheckIcon className="icon-inline" /> No action required. You can view code
                                    intelligence on private repositories synced to Sourcegraph cloud by default.
                                </p>
                            </div>
                            <div className="d-flex align-items-center">
                                <div className="bg-3 rounded-circle p-2 d-flex">
                                    <LockIcon className="icon-inline" />
                                </div>
                                <div className="m-0 ml-3">
                                    <p className="m-0 mb-2">
                                        If you have privacy concerns about your private code, take a look at our{' '}
                                        <a
                                            href="https://docs.sourcegraph.com/integration/browser_extension#privacy"
                                            target="_blank"
                                            rel="noopener"
                                        >
                                            user documentation
                                        </a>{' '}
                                        for a detailed breakdown of what data is passed through Sourcegraph cloud.
                                    </p>
                                    <p className="m-0">
                                        You can also blocklist URLs from passing through Sourcegraph cloud in the
                                        advanced settings.
                                    </p>
                                </div>
                            </div>
                        </div>
                        <div className="col-md-6">
                            <div className="bg-2 rounded p-3 mt-4 mt-md-0 d-flex flex-column justify-content-around">
                                <h3 className="mb-3">Private code in self-hosted Sourcegraph instance</h3>
                                <p>Follow these instructions:</p>
                                <ol className="m-0 d-flex flex-column justify-content-around">
                                    <li>
                                        <strong>Install Sourcegraph</strong> (
                                        <a
                                            href="https://docs.sourcegraph.com/admin/install"
                                            target="_blank"
                                            rel="noopener"
                                        >
                                            visit our docs for instructions
                                        </a>
                                        ).Skip this step if you already have a private Sourcegraph instance.
                                    </li>
                                    <li>
                                        Click the Sourcegraph extension icon in the browser toolbar to{' '}
                                        <a href="./options.html" rel="noopener" target="_blank">
                                            open the settings page
                                        </a>
                                        .
                                    </li>
                                    <li>
                                        Enter the <strong>URL</strong> of your Sourcegraph instance in the second input
                                        field, labeled as <q>Sourcegraph self-hosted</q>.
                                    </li>
                                    <li>
                                        Make sure the status shows <q>Looks good!</q>.
                                    </li>
                                </ol>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            <section className="py-5">
                <div className="container">
                    <h2 className="mb-4">Additional resources</h2>
                    <div className="d-flex w-100 align-items-center">
                        <div className="bg-3 rounded-circle p-2  d-flex">
                            <BookOpenPageVariantIcon className="icon-inline" />
                        </div>
                        <p className="m-0 ml-3">
                            Read the{' '}
                            <a
                                href="https://docs.sourcegraph.com/integration/browser_extension"
                                rel="noopener"
                                target="_blank"
                            >
                                Sourcegraph docs
                            </a>{' '}
                            to learn more about how we respect your privacy, troubleshooting and extension features.
                        </p>
                    </div>
                </div>
            </section>
        </div>
    )
}
