import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Card, CardBody, CardHeader, CardSubtitle, CardText, CardTitle } from '.'

describe('Card', () => {
    it('renders card correctly', () => {
        const { asFragment } = render(
            <Card>
                <CardHeader>Card Header</CardHeader>
                <CardBody>
                    <CardTitle>Card Title</CardTitle>
                    <CardSubtitle>Card Subtitle</CardSubtitle>
                    <CardText>Card Text</CardText>
                </CardBody>
            </Card>
        )

        expect(screen.getByText(/Card Header/)).toHaveClass('cardHeader')
        expect(screen.getByText(/Card Title/)).toHaveClass('cardTitle')
        expect(screen.getByText(/Card Subtitle/)).toHaveClass('cardSubtitle')
        expect(asFragment()).toMatchSnapshot()
    })
})
