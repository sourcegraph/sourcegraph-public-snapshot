import { useCallback, useState } from 'react'

import { mdiChevronDown, mdiChevronLeft } from '@mdi/js'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { H2 } from '..'
import { Button } from '../Button'
import { Input } from '../Form'
import { Icon } from '../Icon'

import { Collapse, CollapseHeader, CollapsePanel } from './Collapse'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Collapse',
    component: Collapse,

    decorators: [decorator],
}

export default config

export const Simple: Story = () => {
    const [isOpened, setIsOpened] = useState(false)

    const handleOpenChange = useCallback((next: boolean) => {
        setIsOpened(next)
    }, [])

    return (
        <div>
            <H2 className="my-3">Controlled collapse</H2>
            <Collapse isOpen={isOpened} onOpenChange={handleOpenChange}>
                <CollapseHeader as={Button} outline={true} focusLocked={true} variant="secondary" className="w-50">
                    Collapsable
                    <Icon aria-hidden={true} svgPath={isOpened ? mdiChevronDown : mdiChevronLeft} className="mr-1" />
                </CollapseHeader>
                <CollapsePanel className="w-50">
                    <Input placeholder="testing this one" />
                </CollapsePanel>
            </Collapse>

            <H2 className="my-3">Uncontrolled collapse</H2>
            <Collapse>
                {({ isOpen }) => (
                    <>
                        <CollapseHeader
                            as={Button}
                            aria-label={isOpen ? 'Expand' : 'Collapse'}
                            outline={true}
                            variant="secondary"
                            className="w-50"
                        >
                            Collapsable
                            <Icon
                                aria-hidden={true}
                                svgPath={isOpen ? mdiChevronDown : mdiChevronLeft}
                                className="mr-1"
                            />
                        </CollapseHeader>
                        <CollapsePanel className="w-50">
                            <Input placeholder="testing this one" />
                        </CollapsePanel>
                    </>
                )}
            </Collapse>

            <H2 className="my-3">Open by default collapse</H2>
            <Collapse openByDefault={true}>
                {({ isOpen }) => (
                    <>
                        <CollapseHeader
                            as={Button}
                            aria-label={isOpen ? 'Expand' : 'Collapse'}
                            outline={true}
                            variant="secondary"
                            className="w-50"
                        >
                            Collapsable
                            <Icon
                                aria-hidden={true}
                                svgPath={isOpen ? mdiChevronDown : mdiChevronLeft}
                                className="mr-1"
                            />
                        </CollapseHeader>
                        <CollapsePanel className="w-50">
                            <Input placeholder="testing this one" />
                        </CollapsePanel>
                    </>
                )}
            </Collapse>

            <H2 className="my-3">Without forced CollapsePanel rendering</H2>
            <Collapse>
                {({ isOpen }) => (
                    <>
                        <CollapseHeader
                            as={Button}
                            aria-label={isOpen ? 'Expand' : 'Collapse'}
                            outline={true}
                            variant="secondary"
                            className="w-50"
                        >
                            Collapsable
                            <Icon
                                aria-hidden={true}
                                svgPath={isOpen ? mdiChevronDown : mdiChevronLeft}
                                className="mr-1"
                            />
                        </CollapseHeader>
                        <CollapsePanel forcedRender={false} className="w-50">
                            <Input placeholder="testing this one" />
                        </CollapsePanel>
                    </>
                )}
            </Collapse>
        </div>
    )
}
