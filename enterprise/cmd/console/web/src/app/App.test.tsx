/**
 * @jest-environment jsdom
 */

import { render, screen } from '@testing-library/react'
import React from 'react'

import { App } from './App'

test('renders page', () => {
    render(<App />)
    const linkElement = screen.getByText(/Hello/)
    expect(linkElement).toBeInTheDocument()
})
