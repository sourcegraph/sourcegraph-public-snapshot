import { mdiMagnify, mdiPlus } from '@mdi/js'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { FeedbackBadge } from '../Feedback'
import { Icon } from '../Icon'
import { Link } from '../Link'

import { PageHeader } from './PageHeader'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/PageHeader',
    component: PageHeader,
    decorators: [decorator],
}

export default config

export const BasicHeader: Story = () => (
    <PageHeader
        path={[{ icon: PuzzleOutlineIcon, text: 'Header' }]}
        actions={
            <Button to={`${location.pathname}/close`} className="mr-1" variant="secondary" as={Link}>
                <Icon aria-hidden={true} svgPath={mdiMagnify} /> Button with icon
            </Button>
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
            <PageHeader.Breadcrumb to="/level-0" icon={PuzzleOutlineIcon} />
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
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
    design: {
        type: 'figma',
        name: 'Figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1485%3A0',
    },
}
