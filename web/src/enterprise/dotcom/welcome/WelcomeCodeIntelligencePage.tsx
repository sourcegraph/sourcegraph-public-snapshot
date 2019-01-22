import H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { CodeIntellifyBlob } from './CodeIntellifyBlob'

const intelligenceSections = [
    {
        title: 'Code browsing',
        paragraph:
            'View open source code, like gorilla/mux, on sourcegraph.com, or deploy your own instance to see public code alongside your private code. See how your codebase changes over time in by browsing through branches, commits, and diffs.',
    },
    {
        title: 'Advanced code intelligence',
        paragraph:
            'Code intelligence makes browsing code easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by language servers based on the open-source Language Server Protocol.',
    },
    {
        title: 'Hover tooltip',
        paragraph:
            'Use the hover tooltip to discover and understand your code faster. Click on a token and then go to its definition, other references, or implementations. Speed through reviews by understanding new code, changed code, and what it affects.',
    },
    {
        title: '',
        paragraph:
            'Code intelligence is powered by language servers based on the open-standard Language Server Protocol (published by Microsoft, with participation from Facebook, Google, Sourcegraph, GitHub, RedHat, Twitter, Salesforce, Eclipse, and others). Visit langserver.org to learn more about the Language Server Protocol, find the latest support for your favorite language, and get involved.',
    },
]

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    location: H.Location
    history: H.History
}

/**
 * The welcome code intelligence page, which describes and demonstrates Sourcegraph's code intelligence
 * functionality.
 */
export class WelcomeCodeIntelligencePage extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="welcome-code-intelligence-page container">
                <h2>Enhanced Code Browsing and Intelligence</h2>
                <h1>Mine your language.</h1>
                <div className="row">
                    <div className="col-md-6">
                        {intelligenceSections.map(({ title, paragraph }, i) => (
                            <div key={i}>
                                <h3>{title}</h3>
                                <p>{paragraph}</p>
                            </div>
                        ))}
                    </div>
                    <div className="col-md-6">
                        <CodeIntellifyBlob
                            {...this.props}
                            startLine={236}
                            endLine={284}
                            parentElement={'.modal-code-intellify'}
                            className={'code-intellify-container-modal'}
                            tooltipClass={'modal-tooltip'}
                            defaultHoverPosition={{ line: 248, character: 11 }}
                        />
                    </div>
                </div>
                <p>
                    Get started with Sourcegraph for free, and get get cross-repository code intelligence, advanced code
                    search, and extensive integrations.
                </p>
                <a className="btn btn-secondary" href="https://docs.sourcegraph.com/#quickstart">
                    Deploy Sourcegraph
                    <ChevronRightIcon className="icon-inline" />
                </a>
                <p>
                    Explore how Sourcegraph's code intelligence can augment and add to your workflow, prepare you for
                    code review, and speed through development.
                </p>
                <a className="btn btn-secondary" href="//about.sourcegraph.com/docs/code-intelligence" target="_blank">
                    Code Intelligence Documentation
                    <ChevronRightIcon className="icon-inline" />
                </a>
            </div>
        )
    }
}
