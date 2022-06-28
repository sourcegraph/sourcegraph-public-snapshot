import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { H1 } from '../Typography'

import { SkipLinksProvider, SkipLink } from './SkipLink'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/SkipLink',
    component: SkipLink,
    decorators: [decorator],
}

export default config

const OtherComponent: React.FunctionComponent = () => (
    <>
        <H1>Hello two!</H1>
        <SkipLink href="#cool" />
    </>
)

export const BasicHeader: Story = () => {
    const [showOther, setShowOther] = useState(false)

    return (
        <SkipLinksProvider>
            <H1>Hello world!</H1>
            <Button variant="secondary" onClick={() => setShowOther(!showOther)}>
                Toggle other component
            </Button>
            <SkipLink href="#main" />
            {showOther && <OtherComponent />}
        </SkipLinksProvider>
    )
}
