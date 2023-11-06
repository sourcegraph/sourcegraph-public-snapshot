import { describe, expect, test } from '@jest/globals'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { Toggle } from './Toggle'

describe('Toggle', () => {
    test('value is false', () => {
        expect(render(<Toggle value={false} />).asFragment()).toMatchSnapshot()
    })

    test('value is true', () => {
        expect(render(<Toggle value={true} />).asFragment()).toMatchSnapshot()
    })

    test('disabled', () => {
        const onToggle = sinon.spy(() => undefined)
        const { asFragment } = render(<Toggle onToggle={onToggle} disabled={true} data-testid="toggle" />)

        userEvent.click(screen.getByTestId('toggle'))
        sinon.assert.notCalled(onToggle)
        expect(asFragment()).toMatchSnapshot()
    })

    test('className', () => expect(render(<Toggle className="c" />).asFragment()).toMatchSnapshot())

    test('aria', () =>
        expect(
            render(
                <Toggle aria-describedby="test-id-1" aria-labelledby="test-id-2" aria-label="test toggle" />
            ).asFragment()
        ).toMatchSnapshot())
})
