import { mdiMagnify, mdiPlus, mdiPuzzleOutline } from '@mdi/js'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '../../stories/BrandedStory'
import { Button } from '../Button'
import { FeedbackBadge } from '../Feedback'
import { Icon } from '../Icon'
import { Link } from '../Link'
import { H1, H2 } from '../Typography'

import { PageHeader } from './PageHeader'

const decorator: Decorator = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageHeader',
    component: PageHeader,
    decorators: [decorator],
}

export default config

export const BasicHeader: StoryFn = () => (
    <>
        <H1>Page Header</H1>
        <H2>Basic</H2>
        <div className="mb-3">
            <PageHeader
                path={[{ icon: mdiPuzzleOutline, text: 'Header' }]}
                actions={
                    <Button to={`${location.pathname}/close`} className="mr-1" variant="secondary" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiMagnify} /> Button with icon
                    </Button>
                }
            />
        </div>
        <H2>Overflowing</H2>
        <div className="mb-3">
            <PageHeader
                path={[
                    {
                        icon: mdiPuzzleOutline,
                        text: 'Call me Ishmael. Some years ago—never mind how long precisely—having little or no money in my purse, and nothing particular to interest me on shore, I thought I would sail about a little and see the watery part of the world.',
                    },
                ]}
            />
        </div>
    </>
)

BasicHeader.storyName = 'Basic header'

BasicHeader.parameters = {
    design: {
        type: 'figma',
        name: 'Figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1485%3A0',
    },
}

export const ComplexHeader: StoryFn = () => (
    <PageHeader
        annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
        byline={
            <>
                Created by <Link to="/page">user</Link> 3 months ago
            </>
        }
        description="Enter the description for your section here. This is useful on list and create pages."
        actions={
            <div className="d-flex">
                <Button as={Link} to="/page" variant="secondary" className="mr-2">
                    Secondary
                </Button>
                <Button as={Link} to="/page" variant="primary" className="text-nowrap">
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Create
                </Button>
            </div>
        }
    >
        <PageHeader.Heading as="h2" styleAs="h1">
            <PageHeader.Breadcrumb to="/level-0" icon={mdiPuzzleOutline} />
            <PageHeader.Breadcrumb to="/level-1">Level 1</PageHeader.Breadcrumb>
            <PageHeader.Breadcrumb>Level 2</PageHeader.Breadcrumb>
            <PageHeader.Breadcrumb>Level 3</PageHeader.Breadcrumb>
            <PageHeader.Breadcrumb>Level 4</PageHeader.Breadcrumb>
            <PageHeader.Breadcrumb>Level 5</PageHeader.Breadcrumb>
        </PageHeader.Heading>
    </PageHeader>
)

ComplexHeader.storyName = 'Complex header'

ComplexHeader.parameters = {
    design: {
        type: 'figma',
        name: 'Figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1485%3A0',
    },
}
