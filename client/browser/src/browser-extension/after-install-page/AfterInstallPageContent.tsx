import React from 'react'
import LockIcon from 'mdi-react/LockIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import { PhabricatorIcon } from '../../../../shared/src/components/icons'
import { SourcegraphIcon } from '../../shared/components/SourcegraphIcon'

export const AfterInstallPageContent: React.FunctionComponent = () => (
    <div className="web-content after-install-page-content">
        <SourcegraphIcon className="after-install-page-content__sourcegraph-logo" />

        <div className="container">
            <h1 className="mt-5">🎉 You’ve just installed the Sourcegraph browser extension!</h1>
            <p className="lead mb-0">We’ve gathered the most important information that will get your started:</p>
        </div>

        <section className="border-bottom py-5">
            <div className="container">
                <h2 className="mb-4">How do I use the extension?</h2>
                <div className="row">
                    <div className="col-sm-6">
                        <h3>Code intelligence on your code host</h3>
                        <p className="m-0">
                            Sourcegraph browser extension adds code intelligence to files and diffs on GitHub, GitHub
                            Enterprise, GitLab, Phabricator, and Bitbucket Server.
                        </p>
                        {/* Video placeholder */}
                    </div>
                    <div className="col-sm-6">
                        <h3>Search shortcut in the URL location bar</h3>
                        <p className="m-0">
                            Type <code>src</code>
                            <kbd>space</kbd> in the address bar of your browser to search for queries on Sourcegraph.
                        </p>
                        {/* Video placeholder */}
                    </div>
                </div>
            </div>
        </section>

        <section className="border-bottom py-5">
            <div className="container">
                <h2 className="mb-4">Make it work on your codehost</h2>
                <div className="row">
                    <div className="col-sm-6">
                        <div className="bg-2 rounded p-3 mb-3">
                            <h3 className="mb-3 after-install-page-content__code-host-titles">
                                <GithubIcon className="icon-inline after-install-page-content__code-host-logo" />{' '}
                                github.com
                            </h3>
                            <p className="m-0">
                                <CheckIcon className="icon-inline" /> No action required. Your extension works here by
                                default.
                            </p>
                        </div>
                        <div className="bg-2 rounded p-3">
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
                    <div className="col-sm-6">{/* Video placeholder */}</div>
                </div>
            </div>
        </section>

        <section className="border-bottom py-5">
            <div className="container">
                <h2 className="mb-4">Make it work for private code</h2>
                <div className="row">
                    <div className="col-sm-6">
                        <p>By default, the browser extension works only for public code.</p>
                        <div className="d-flex w-100 align-items-center">
                            <div className="bg-3 rounded-circle p-2">
                                <LockIcon className="icon-inline" />
                            </div>
                            <p className="flex-grow-1 m-0 ml-3">
                                To use the browser extension with your private repositories, you need to set up a{' '}
                                <strong>private Sourcegraph instance</strong> and connect the extension to it.
                            </p>
                        </div>
                        <div className="bg-2 rounded p-3 mt-4">
                            <p>Follow these instructions:</p>
                            <ol className="m-0">
                                <li>
                                    <strong>Install Sourcegraph</strong> (
                                    <a href="https://docs.sourcegraph.com/admin/install" target="_blank" rel="noopener">
                                        visit our docs for instructions
                                    </a>
                                    ). Skip this step if you already have a private Sourcegraph instance.
                                </li>
                                <li>
                                    Click the Sourcegraph extension icon in the browser toolbar to{' '}
                                    <a href="./options.html" rel="noopener" target="_blank">
                                        open the settings page
                                    </a>
                                    .
                                </li>
                                <li>
                                    Enter the <strong>URL</strong> (including the protocol) of your Sourcegraph instance
                                    (such as <q>https://sourcegraph.example.com</q>).
                                </li>
                                <li>Make sure a green checkmark appears in the input field.</li>
                            </ol>
                        </div>
                    </div>
                    <div className="col-sm-6">{/* Video placeholder */}</div>
                </div>
            </div>
        </section>

        <section className="py-5">
            <div className="container">
                <h2 className="mb-4">Additional resources</h2>
                <div className="d-flex w-100 align-items-center">
                    <div className="bg-3 rounded-circle p-2">
                        <BookOpenPageVariantIcon className="icon-inline" />
                    </div>
                    <p className="flex-grow-1 m-0 ml-3">
                        Read the{' '}
                        <a
                            href="https://docs.sourcegraph.com/integration/browser_extension"
                            rel="noopener"
                            target="_blank"
                        >
                            Sourcegraph Documentation
                        </a>{' '}
                        to learn more about how we respect your privacy, troubleshooting and extension features.
                    </p>
                </div>
            </div>
        </section>
    </div>
)
