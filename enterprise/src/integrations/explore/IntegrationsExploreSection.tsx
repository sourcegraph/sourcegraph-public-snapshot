import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'

const DATA: { title: string; description: string; url: string; className: string }[] = [
    {
        title: 'Browser extension',
        description: 'Go-to-definition and hovers in your code host and reviews.',
        url: 'https://docs.sourcegraph.com/integration/browser_extension',
        className: 'bg-primary',
    },
    {
        title: 'src CLI',
        description: 'Search code from your terminal.',
        url: 'https://github.com/sourcegraph/src-cli',
        className: 'bg-warning',
    },
    {
        title: 'Editor extensions',
        description: 'Jump to Sourcegraph from your editor.',
        url: 'https://docs.sourcegraph.com/integration/editor',
        className: 'bg-info',
    },
]

/**
 * An explore section that shows integrations.
 */
export class IntegrationsExploreSection extends React.PureComponent {
    public render(): JSX.Element | null {
        return (
            <div className="integrations-explore-section">
                <h2>Popular integrations</h2>
                <div className="card-deck">
                    {DATA.map(({ title, description, url, className }, i) => (
                        <a
                            key={i}
                            href={url}
                            target="_blank"
                            className={`card border-white card-link text-white ${className}`}
                        >
                            <div className="card-body">
                                <h2 className="card-title h6 font-weight-bold">{title}</h2>
                                <p className="card-text">{description}</p>
                            </div>
                        </a>
                    ))}
                </div>
                <div className="text-right mt-3">
                    <a href="https://docs.sourcegraph.com/integration" target="_blank">
                        View all integrations<ChevronRightIcon className="icon-inline" />
                    </a>
                </div>
            </div>
        )
    }
}
