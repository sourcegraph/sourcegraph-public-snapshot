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
                iconUrl: 'https://about.sourcegraph.com/integrations/chrome.svg',
                text: 'Chrome',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/firefox.svg',
                text: 'Firefox',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
        ],
    },

    {
        title: 'Code host integrations',
        paragraph:
            'The Sourcegraph browser extension will add go-to-definition, find-references, hover tooltips, and code search to all files and diffs on supported code hosts. The extension will also add code intelligence and code search to public repositories. ',
        buttons: [
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/gitlab.svg',
                text: 'GitLab',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/github.png',
                text: 'GitHub',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/phabricator.png',
                text: 'Phabricator',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
        ],
    },
    {
        title: 'Editor extensions',
        paragraph:
            'Our editor extensions let you quickly jump to files and search code on your Sourcegraph instance from your editor. Seamlessly jump for development to review without missing a step.',
        buttons: [
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/atom.svg',
                text: 'Atom',
                link: 'https://atom.io/packages/sourcegraph',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/jetbrains.svg',
                text: 'IntelliJ',
                link: 'https://plugins.jetbrains.com/plugin/9682-sourcegraph',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/sublime.svg',
                text: 'Sublime',
                link: 'https://github.com/sourcegraph/sourcegraph-sublime',
            },
            {
                iconUrl: 'https://about.sourcegraph.com/integrations/vscode.svg',
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
                        {buttons.map(({ text, iconUrl, link }, i) => (
                            <a key={i} className="btn btn-secondary btn-integrations" href={link}>
                                {/* tslint:disable-next-line:jsx-ban-props */}
                                <img src={iconUrl} style={{ maxHeight: '24px' }} className="mr-1" />
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
                <a href="https://about.sourcegraph.com/docs/integrations" target="_blank">
                    Integrations Documentation
                    <ChevronRightIcon className="icon-inline" />
                </a>
            </div>
        )
    }
}
