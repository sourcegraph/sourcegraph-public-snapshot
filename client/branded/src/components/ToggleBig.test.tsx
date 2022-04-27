import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { ToggleBig } from './ToggleBig'

describe('ToggleBig', () => {
    test('value is false', () => {
        expect(render(<ToggleBig value={false} />).asFragment()).toMatchSnapshot()
    })

    test('value is true', () => {
        expect(render(<ToggleBig value={true} />).asFragment()).toMatchSnapshot()
    })

    test('disabled', () => {
        const onToggle = sinon.spy(() => undefined)
        const { asFragment } = render(<ToggleBig onToggle={onToggle} disabled={true} />)
        userEvent.click(screen.getByRole('switch'))

        sinon.assert.notCalled(onToggle)
        expect(asFragment()).toMatchSnapshot()
    })

    test('className', () => expect(render(<ToggleBig className="c" />).asFragment()).toMatchSnapshot())
})
