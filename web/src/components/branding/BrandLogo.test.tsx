import React from 'react'
import renderer from 'react-test-renderer'
import { BrandLogo } from './BrandLogo'

describe('BrandLogo', () => {
    test('renders light sourcegraph logo', () =>
        expect(renderer.create(<BrandLogo assetsRoot="/assets" isLightTheme={true} />).toJSON()).toMatchSnapshot())

    test('renders dark sourcegraph logo', () =>
        expect(renderer.create(<BrandLogo assetsRoot="/assets" isLightTheme={false} />).toJSON()).toMatchSnapshot())

    test('renders light custom branding logo', () =>
        expect(
            renderer
                .create(
                    <BrandLogo
                        branding={{ brandName: 'b', light: { logo: 'l' } }}
                        assetsRoot="/assets"
                        isLightTheme={true}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())

    test('renders dark custom branding logo', () =>
        expect(
            renderer
                .create(
                    <BrandLogo
                        branding={{ brandName: 'b', dark: { logo: 'l' } }}
                        assetsRoot="/assets"
                        isLightTheme={false}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
