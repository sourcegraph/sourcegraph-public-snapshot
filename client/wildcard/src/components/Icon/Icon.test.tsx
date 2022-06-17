import { mdiCheck } from '@mdi/js'
import { render } from '@testing-library/react'

import { SourcegraphIcon } from '../SourcegraphIcon'

import { IconV2 } from './Icon'

describe('Icon', () => {
    describe('custom icons', () => {
        it('renders a simple inline icon correctly', () => {
            const { asFragment } = render(<IconV2 as={SourcegraphIcon} aria-hidden={true} />)
            expect(asFragment()).toMatchSnapshot()
        })
        it('renders a medium icon correctly', () => {
            const { asFragment } = render(<IconV2 as={SourcegraphIcon} size="md" aria-label="Sourcegraph logo" />)
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('mdi icons', () => {
        it('renders a simple inline icon correctly', () => {
            const { asFragment } = render(<IconV2 svgPath={mdiCheck} aria-hidden={true} />)
            expect(asFragment()).toMatchSnapshot()
        })
        it('renders a medium icon correctly', () => {
            const { asFragment } = render(<IconV2 svgPath={mdiCheck} size="md" aria-label="Checkmark" />)
            expect(asFragment()).toMatchSnapshot()
        })
    })
})
