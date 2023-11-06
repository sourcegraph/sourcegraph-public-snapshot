import { mdiClose } from '@mdi/js'
import { render } from '@testing-library/react'
import CloseIcon from 'mdi-react/CloseIcon'
import { describe, expect, it } from 'vitest'

import { SourcegraphIcon } from '../SourcegraphIcon'

import { Icon } from './Icon'

describe('Icon', () => {
    describe('custom icons', () => {
        it('renders a simple inline icon correctly', () => {
            const { asFragment } = render(<Icon as={SourcegraphIcon} aria-hidden={true} />)
            expect(asFragment()).toMatchSnapshot()
        })

        it('renders a medium icon correctly', () => {
            const { asFragment } = render(<Icon as={SourcegraphIcon} size="md" aria-label="Sourcegraph logo" />)
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('legacy mdi-react icons', () => {
        it('renders a simple inline icon correctly', () => {
            const { asFragment } = render(<Icon as={CloseIcon} aria-hidden={true} />)
            expect(asFragment()).toMatchSnapshot()
        })

        it('renders a medium icon correctly', () => {
            const { asFragment } = render(<Icon as={CloseIcon} size="md" aria-label="Checkmark" />)
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('new @mdi/js icons', () => {
        it('renders a simple inline icon correctly', () => {
            const { asFragment } = render(<Icon svgPath={mdiClose} aria-hidden={true} />)
            expect(asFragment()).toMatchSnapshot()
        })
        it('renders a medium icon correctly', () => {
            const { asFragment } = render(<Icon svgPath={mdiClose} size="md" aria-label="Checkmark" />)
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('styled as icons', () => {
        it('renders a simple span', () => {
            const { asFragment } = render(
                <Icon as="span" aria-label="Smile">
                    :)
                </Icon>
            )
            expect(asFragment()).toMatchSnapshot()
        })
    })
})
