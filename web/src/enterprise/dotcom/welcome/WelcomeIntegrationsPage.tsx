import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'

const integrationsSections = [
    {
        title: 'Connect across your development workflow.',
        paragraph:
            'Sourcegraph has powerful integrations for every step of development. From planning with code discussion, development with Sourcegraph and IDE extensions, to review in PRs and Issues. Use Sourcegraph integrations to get code intelligence at every step of your workflow.',
        buttons: [],
    },
    {
        title: 'Browser extensions',
        paragraph:
            'Code intelligence makes browsing code easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by language servers based on the open-source Language Server Protocol.',
        buttons: [
            {
                id: 'btn-chrome',
                text: 'Chrome',
                link: 'https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack',
            },
        ],
    },

    {
        title: 'Code host integrations',
        paragraph:
            'The Sourcegraph browser extension will add go-to-definition, find-references, hover tooltips, and code search to all files and diffs on supported code hosts. The extension will also add code intelligence and code search to public repositories. ',
        buttons: [
            { id: 'btn-gitlab', text: 'GitLab', link: 'https://docs.sourcegraph.com/integration/browser_extension' },
            { id: 'btn-github', text: 'GitHub', link: 'https://docs.sourcegraph.com/integration/browser_extension' },
            {
                id: 'btn-phabricator',
                text: 'Phabricator',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
        ],
    },
    {
        title: 'IDE extensions',
        paragraph:
            'Our editor plugins let you quickly jump to files and search code on your Sourcegraph instance from your editor. Seamlessly jump for development to review without missing a step.',
        buttons: [
            { id: 'btn-atom', text: 'Atom', link: 'https://atom.io/packages/sourcegraph' },
            { id: 'btn-intellij', text: 'IntelliJ', link: 'https://plugins.jetbrains.com/plugin/9682-sourcegraph' },
            {
                id: 'btn-sublime',
                text: 'Sublime',
                link: 'https://github.com/sourcegraph/sourcegraph-sublime',
            },
            {
                id: 'btn-vscode',
                text: 'VS Code',
                link: 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph',
            },
        ],
    },
]

interface Props {}

/**
 * The welcome integrations page, which describes and demonstrates Sourcegraph's integrations.
 */
export class WelcomeIntegrationsPage extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="welcome-integrations-page">
                <h2>Integrations</h2>
                <h1>Get it. Together.</h1>
                {integrationsSections.map(({ title, paragraph, buttons }, i) => (
                    <div key={i}>
                        <h3>{title}</h3>
                        <p>{paragraph}</p>
                        {buttons.map(({ text, id, link }, j) => (
                            <a
                                key={`integrations-buttons-${j}`}
                                className={`btn btn-secondary btn-integrations  ${id}`}
                                href={`${link}`}
                            >
                                <span className="logo-icon" />
                                {text}
                            </a>
                        ))}
                    </div>
                ))}
                <p>
                    Get started with Sourcegraph for free, and get get cross-repository code intelligence, advanced code
                    search, and extensive integrations.
                </p>
                <a href="https://docs.sourcegraph.com/#quickstart">
                    Deploy Sourcegraph
                    <ChevronRightIcon className="icon-inline" />
                </a>
                <p>
                    Explore all of Sourcegraph's integrations and see how you can get cross-repository code intelligence
                    on your favorite code host and editor.
                </p>
                <a href="//about.sourcegraph.com/docs/integrations" target="_blank">
                    Integrations Documentation
                    <ChevronRightIcon className="icon-inline" />
                </a>
            </div>
        )
    }
}
