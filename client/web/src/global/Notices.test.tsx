import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { Notices } from './Notices'

describe('Notices', () => {
    test('shows notices for location', () =>
        expect(renderWithBrandedContext(<Notices location="home" />).asFragment()).toMatchSnapshot())

    test('no notices', () =>
        expect(renderWithBrandedContext(<Notices location="home" />).asFragment()).toMatchSnapshot())
})
