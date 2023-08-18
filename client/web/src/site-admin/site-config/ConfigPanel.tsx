import { useState, type FC } from 'react'

import { mdiChevronDown, mdiChevronRight, mdiLink } from '@mdi/js'
import { noop } from 'lodash'
import { HashLink } from 'react-router-hash-link'

import { Collapse, CollapseHeader, CollapsePanel, Container, Icon } from '@sourcegraph/wildcard'

interface Props {
    id: string
    title: string
    className?: string
    children: React.ReactNode
    defaultOpen?: boolean
}

export const ConfigPanel: FC<Props> = ({ className, id, title, children, defaultOpen }) => {
    // TODO: do not open by default if config is OK
    const [open, setOpen] = useState(defaultOpen ?? true)

    const toggleOpen = (): void => setOpen(prev => !prev)

    return (
        <Container className={className}>
            <Collapse isOpen={open} onOpenChange={noop}>
                <CollapseHeader as="h3" id={id}>
                    <span role="button" tabIndex={0} onClick={toggleOpen} onKeyDown={toggleOpen}>
                        {open ? (
                            <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                        ) : (
                            <Icon aria-hidden={true} svgPath={mdiChevronRight} />
                        )}{' '}
                        {title}
                    </span>{' '}
                    <HashLink smooth={true} to={`#${id}`}>
                        <Icon aria-label="link icon" svgPath={mdiLink} />
                    </HashLink>
                </CollapseHeader>
                <CollapsePanel>{children}</CollapsePanel>
            </Collapse>
        </Container>
    )
}
