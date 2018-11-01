import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'

const DATA: { title: string; description: string; url: string; backgroundImage: string }[] = [
    {
        title: 'Browser extension',
        description: 'Go-to-definition and hovers in your code host and reviews.',
        url: 'https://docs.sourcegraph.com/integration/browser_extension',
        backgroundImage:
            'linear-gradient(96deg, #397b9c, #b553af 46%, #bb5525), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
    },
    {
        title: 'src CLI',
        description: 'Search code from your terminal.',
        url: 'https://github.com/sourcegraph/src-cli',
        backgroundImage:
            'linear-gradient(100deg, #1b82e8, #023dc9), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
    },
    {
        title: 'Editor extensions',
        description: 'Jump to Sourcegraph from your editor.',
        url: 'https://docs.sourcegraph.com/integration/editor',
        backgroundImage:
            'linear-gradient(100deg, #36334c, #2b5897), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
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
                <div className="row">
                    {DATA.map(({ title, description, url, backgroundImage }, i) => (
                        <div key={i} className="col-md-4 mb-2 mb-md-0">
                            <a
                                href={url}
                                target="_blank"
                                className="card rounded border-white card-link text-white"
                                // tslint:disable-next-line:jsx-ban-props
                                style={{ backgroundImage }}
                            >
                                <div className="card-body">
                                    <h2 className="card-title h6 font-weight-bold mb-0">{title}</h2>
                                    <p className="card-text">{description}</p>
                                </div>
                            </a>
                        </div>
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
