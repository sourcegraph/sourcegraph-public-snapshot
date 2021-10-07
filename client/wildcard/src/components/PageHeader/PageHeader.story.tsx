import { DecoratorFn, Meta, Story } from '@storybook/react'
import PlusIcon from 'mdi-react/PlusIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { FeedbackBadge } from '@sourcegraph/web/src/components/FeedbackBadge'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { PageHeader } from './PageHeader'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageHeader',
    decorators: [decorator],
}

export default config

export const BasicHeader: Story = () => (
    <PageHeader
        path={[{ icon: PuzzleOutlineIcon, text: 'Header' }]}
        actions={
            <Link to={`${location.pathname}/close`} className="btn btn-secondary mr-1">
                <SearchIcon className="icon-inline" /> Button with icon
            </Link>
        }
    />
)

BasicHeader.storyName = 'Basic header'

BasicHeader.parameters = {
    design: {
        type: 'figma',
        name: 'Figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1485%3A0',
    },
}

export const ComplexHeader: Story = () => (
    <PageHeader
        annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
        path={[{ to: '/level-0', icon: PuzzleOutlineIcon }, { to: '/level-1', text: 'Level 1' }, { text: 'Level 2' }]}
        byline={
            <>
                Created by <Link to="/page">user</Link> 3 months ago
            </>
        }
        description="Enter the description for your section here. This is useful on list and create pages."
        actions={
            <div className="d-flex">
                <Link to="/page" className="btn btn-secondary mr-2">
                    Secondary
                </Link>
                <Link to="/page" className="btn btn-primary text-nowrap">
                    <PlusIcon className="icon-inline" /> Create
                </Link>
            </div>
        }
    />
)

ComplexHeader.storyName = 'Complex header'

ComplexHeader.parameters = {
    design: {
        type: 'figma',
        name: 'Figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1485%3A0',
    },
}
