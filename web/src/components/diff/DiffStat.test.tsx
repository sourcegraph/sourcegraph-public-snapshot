import React from 'react'
import { DiffStat } from './DiffStat'
import { mount } from 'enzyme'

describe('DiffStat', () => {
    test('standard', () => expect(mount(<DiffStat added={1} changed={2} deleted={3} />).children()).toMatchSnapshot())

    test('expandedCounts', () =>
        expect(
            mount(<DiffStat added={1} changed={2} deleted={3} expandedCounts={true} />).children()
        ).toMatchSnapshot())
})
