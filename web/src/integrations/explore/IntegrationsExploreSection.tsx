import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React from 'react'
import { Link } from 'react-router-dom'

const DATA: { title: string; description: string; url: string }[] = [
    {
        title: 'Browser extension',
        description: 'Go-to-definition and hovers in your code host and reviews.',
        url: '/help/integration/browser_extension',
    },
    {
        title: 'Sourcegraph CLI',
        description: 'Search code from your terminal.',
        url: 'https://github.com/sourcegraph/src-cli',
    },
    {
        title: 'Editor extensions',
        description: 'Jump to Sourcegraph from your editor.',
        url: '/help/integration/editor',
    },
]

interface Props {}

/**
 * An explore section that shows integrations.
 */
export const IntegrationsExploreSection: React.FunctionComponent<Props> = () => (
    <div className="card">
        <h3 className="card-header">Popular integrations</h3>
        <div className="list-group list-group-flush">
            {DATA.map(({ title, description, url }, index) => (
                <a
                    key={index}
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="list-group-item list-group-item-action"
                >
                    <h4 className="mb-0">{title}</h4>
                    <small>{description}</small> <OpenInNewIcon className="icon-inline small" />
                </a>
            ))}
        </div>
        <div className="card-footer">
            <Link to="/help/integration" target="_blank">
                View all integrations
                <ChevronRightIcon className="icon-inline" />
            </Link>
        </div>
    </div>
)
