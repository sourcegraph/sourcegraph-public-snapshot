import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { Link } from '../Link'
import { H1, H2 } from '../Typography'

import { SkipLinkProvider } from './SkipLinkProvider'

import { SkipLink } from '.'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/SkipLink',
    component: SkipLink,
    decorators: [decorator],
}

export default config

const NavBar: React.FunctionComponent = () => (
    <nav style={{ display: 'flex', marginBottom: '1rem' }}>
        <Button variant="primary">Action to skip 1</Button>
        <Button variant="secondary">Action to skip 2</Button>
        <Button variant="secondary">Action to skip 3</Button>
    </nav>
)

const OtherComponent: React.FunctionComponent = () => (
    <>
        <SkipLink id="skip-to-other-content" name="Skip to other content" />
        <H2>I am a different component</H2>
        <Link to="https://sourcegraph.com">Link to somewhere</Link>
    </>
)

export const BasicUsage: Story = () => {
    const [showOther, setShowOther] = useState(false)

    return (
        <SkipLinkProvider>
            <NavBar />
            <SkipLink id="skip-to-content" name="Skip to content" />
            <H1>Hello world!</H1>
            <Button variant="secondary" className="mb-3" onClick={() => setShowOther(!showOther)}>
                Toggle a new component
            </Button>
            {showOther && <OtherComponent />}
        </SkipLinkProvider>
    )
}

export const UsingCustomIdentifier: Story = () => (
    <SkipLinkProvider>
        <NavBar />
        <SkipLink id="skip-to-content" name="Skip to content" renderAnchor={false} />
        <H1>Hello world!</H1>
        <Link id="skip-to-content" to="https://sourcegraph.com">
            Link to somewhere
        </Link>
    </SkipLinkProvider>
)
