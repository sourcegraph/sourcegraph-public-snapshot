import React from 'react'
import { BrandLogo } from './BrandLogo'
import { mount } from 'enzyme'

describe('BrandLogo', () => {
    test('renders light sourcegraph logo', () =>
        expect(mount(<BrandLogo assetsRoot="/assets" isLightTheme={true} />).children()).toMatchSnapshot())

    test('renders dark sourcegraph logo', () =>
        expect(mount(<BrandLogo assetsRoot="/assets" isLightTheme={false} />).children()).toMatchSnapshot())

    test('renders light custom branding logo', () =>
        expect(
            mount(
                <BrandLogo
                    branding={{ brandName: 'b', light: { logo: 'l' } }}
                    assetsRoot="/assets"
                    isLightTheme={true}
                />
            ).children()
        ).toMatchSnapshot())

    test('renders dark custom branding logo', () =>
        expect(
            mount(
                <BrandLogo
                    branding={{ brandName: 'b', dark: { logo: 'l' } }}
                    assetsRoot="/assets"
                    isLightTheme={false}
                />
            ).children()
        ).toMatchSnapshot())
})
