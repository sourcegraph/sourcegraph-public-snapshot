import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { describe, expect, test } from 'vitest'

import { renderWithBrandedContext } from '../../testing'

import { ButtonLink } from './ButtonLink'

describe('<ButtonLink />', () => {
    test('renders correctly btn classes', () => {
        const { asFragment } = renderWithBrandedContext(
            <ButtonLink to="http://example.com" variant="secondary" size="lg">
                Button link
            </ButtonLink>
        )
        expect(asFragment()).toMatchSnapshot()
    })
    test('renders correctly `disabled`', () => {
        const { asFragment } = renderWithBrandedContext(
            <ButtonLink to="http://example.com" variant="secondary" size="lg" disabled={true}>
                Button link
            </ButtonLink>
        )
        expect(asFragment()).toMatchSnapshot()
    })
    test('renders correctly empty `to`', () => {
        const { asFragment } = renderWithBrandedContext(
            <ButtonLink to={undefined} variant="secondary" size="lg">
                Button link
            </ButtonLink>
        )
        expect(asFragment()).toMatchSnapshot()
    })
    test('renders correctly anchor attributes', () => {
        const { asFragment } = renderWithBrandedContext(
            <ButtonLink
                to="https://sourcegraph.com"
                variant="secondary"
                size="lg"
                target="_blank"
                rel="noopener noreferrer"
                data-pressed="true"
            >
                Button link
            </ButtonLink>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('Should trigger onSelect', () => {
        const onSelect = sinon.stub()

        renderWithBrandedContext(
            <ButtonLink
                to=""
                variant="secondary"
                size="lg"
                target="_blank"
                rel="noopener noreferrer"
                data-pressed="true"
                onClick={onSelect}
                data-testid="button-link"
            >
                Button link
            </ButtonLink>
        )

        userEvent.click(screen.getByTestId('button-link'))

        sinon.assert.calledOnce(onSelect)
    })
})
