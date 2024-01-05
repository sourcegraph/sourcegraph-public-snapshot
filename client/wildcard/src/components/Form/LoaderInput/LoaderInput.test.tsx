import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Input } from '../Input'

import { LoaderInput } from './LoaderInput'

describe('LoaderInput', () => {
    it('should render a loading spinner when loading prop is true', () => {
        expect(
            render(
                <LoaderInput loading={true}>
                    <Input />
                </LoaderInput>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should not render a loading spinner when loading prop is false', () => {
        expect(
            render(
                <LoaderInput loading={false}>
                    <Input />
                </LoaderInput>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
