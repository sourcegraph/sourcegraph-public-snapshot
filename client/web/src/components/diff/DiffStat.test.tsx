import { render } from '@testing-library/react'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

describe('DiffStat', () => {
    test('standard', () =>
        expect(render(<DiffStat added={1} changed={2} deleted={3} />).asFragment()).toMatchSnapshot())

    test('expandedCounts', () =>
        expect(
            render(<DiffStat added={1} changed={2} deleted={3} expandedCounts={true} />).asFragment()
        ).toMatchSnapshot())
})

test('DiffStatSquares', () =>
    expect(render(<DiffStatSquares added={1} changed={2} deleted={3} />).asFragment()).toMatchSnapshot())

test('DiffStatStack', () =>
    expect(render(<DiffStatStack added={1} changed={2} deleted={3} />).asFragment()).toMatchSnapshot())
