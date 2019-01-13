import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React from 'react'
import { Link } from 'react-router-dom'

const DATA: { title: string; description: string; url: string; backgroundImage: string }[] = [
    {
        title: 'Browser extension',
        description: 'Go-to-definition and hovers in your code host and reviews.',
        url: '/help/integration/browser_extension',
        backgroundImage:
            'linear-gradient(96deg, #397b9c, #b553af 46%, #bb5525), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
    },
    {
        title: 'Sourcegraph CLI',
        description: 'Search code from your terminal.',
        url: 'https://github.com/sourcegraph/src-cli',
        backgroundImage:
            'linear-gradient(100deg, #1b82e8, #023dc9), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
    },
    {
        title: 'Editor extensions',
        description: 'Jump to Sourcegraph from your editor.',
        url: '/help/integration/editor',
        backgroundImage:
            'linear-gradient(100deg, #36334c, #2b5897), linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.1))',
    },
]

interface Props {}

/**
 * An explore section that shows integrations.
 */
export const IntegrationsExploreSection: React.FunctionComponent<Props> = () => (
    <div className="integrations-explore-section">
        <h2 className="integrations-explore-section__section-title">Popular integrations</h2>
        <div className="integrations-explore-section__row">
            {DATA.map(({ title, description, url, backgroundImage }, i) => (
                <div key={i} className="integrations-explore-section__card">
                    <a
                        href={url}
                        target="_blank"
                        className="integrations-explore-section__card__content"
                        // tslint:disable-next-line:jsx-ban-props
                        // style={{ backgroundImage }}
                    >
                        <div className="integrations-explore-section__card__content__body">
                            <div className="integrations-explore-section__card__content__body__content">
                                <h2 className="integrations-explore-section__card__content__body__title">{title}</h2>
                                <p className="integrations-explore-section__card__content__body__text">{description}</p>
                            </div>
                            <div className="integrations-explore-section__card__content__icon">
                                <OpenInNewIcon className="" />
                            </div>
                        </div>
                    </a>
                </div>
            ))}
        </div>
        <div className="text-right mt-2">
            <Link to="/help/integration" target="_blank">
                View all integrations
                <ChevronRightIcon className="icon-inline" />
            </Link>
        </div>
    </div>
)
