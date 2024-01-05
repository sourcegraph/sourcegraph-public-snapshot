import { render } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

describe('DiffStat', () => {
    test('standard', () => expect(render(<DiffStat added={3} deleted={5} />).asFragment()).toMatchSnapshot())

    test('expandedCounts', () =>
        expect(render(<DiffStat added={3} deleted={5} expandedCounts={true} />).asFragment()).toMatchSnapshot())
})

test('DiffStatSquares', () => expect(render(<DiffStatSquares added={3} deleted={5} />).asFragment()).toMatchSnapshot())

test('DiffStatStack', () => expect(render(<DiffStatStack added={3} deleted={5} />).asFragment()).toMatchSnapshot())
