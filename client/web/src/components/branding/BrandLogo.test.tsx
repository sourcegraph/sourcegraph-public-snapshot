import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'

import { BrandLogo } from './BrandLogo'

describe('BrandLogo', () => {
    test('renders light sourcegraph logo', () =>
        expect(
            render(<BrandLogo assetsRoot="/assets" isLightTheme={true} variant="logo" />).asFragment()
        ).toMatchSnapshot())

    test('renders light sourcegraph symbol', () =>
        expect(
            render(<BrandLogo assetsRoot="/assets" isLightTheme={true} variant="symbol" />).asFragment()
        ).toMatchSnapshot())

    test('renders dark sourcegraph logo', () =>
        expect(
            render(<BrandLogo assetsRoot="/assets" isLightTheme={false} variant="logo" />).asFragment()
        ).toMatchSnapshot())

    test('renders dark sourcegraph symbol', () =>
        expect(
            render(<BrandLogo assetsRoot="/assets" isLightTheme={false} variant="symbol" />).asFragment()
        ).toMatchSnapshot())

    test('renders light custom branding logo', () =>
        expect(
            render(
                <BrandLogo
                    branding={{ brandName: 'b', light: { logo: 'l' } }}
                    assetsRoot="/assets"
                    isLightTheme={true}
                    variant="logo"
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('renders light custom branding symbol', () =>
        expect(
            render(
                <BrandLogo
                    branding={{ brandName: 'b', light: { symbol: 's' } }}
                    assetsRoot="/assets"
                    isLightTheme={true}
                    variant="symbol"
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('renders dark custom branding logo', () =>
        expect(
            render(
                <BrandLogo
                    branding={{ brandName: 'b', dark: { logo: 'l' } }}
                    assetsRoot="/assets"
                    isLightTheme={false}
                    variant="logo"
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('renders dark custom branding symbol', () =>
        expect(
            render(
                <BrandLogo
                    branding={{ brandName: 'b', dark: { symbol: 's' } }}
                    assetsRoot="/assets"
                    isLightTheme={false}
                    variant="symbol"
                />
            ).asFragment()
        ).toMatchSnapshot())
})
