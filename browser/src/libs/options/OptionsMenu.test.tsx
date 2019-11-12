import { noop } from 'lodash'
import React from 'react'
import renderer from 'react-test-renderer'
import { cleanup, fireEvent, render } from '@testing-library/react'
import sinon from 'sinon'

import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { OptionsMenu, OptionsMenuProps } from './OptionsMenu'

describe('ServerURLForm', () => {
    afterAll(cleanup)

    const stubs: OptionsMenuProps = {
        status: 'connected',
        version: '0.0.0',
        urlHasPermissions: true,
        sourcegraphURL: DEFAULT_SOURCEGRAPH_URL,
        requestPermissions: noop,
        onURLChange: noop,
        onURLSubmit: noop,
        isActivated: true,
        toggleFeatureFlag: noop,
        onToggleActivationClick: noop,
        onSettingsClick: noop,
    }

    test('renders a default state', () => {
        expect(renderer.create(<OptionsMenu {...stubs} />)).toMatchSnapshot()
    })

    test('renders the current tab permissions alert', () => {
        expect(
            renderer.create(
                <OptionsMenu
                    {...stubs}
                    currentTabStatus={{ host: 'gitlab.com', protocol: 'http', hasPermissions: false }}
                />
            )
        ).toMatchSnapshot()
    })

    test("doesn't render the permissions alert on chrome://extensions", () => {
        expect(
            renderer.create(
                <OptionsMenu
                    {...stubs}
                    currentTabStatus={{ host: 'extensions', protocol: 'chrome:', hasPermissions: false }}
                />
            )
        ).toMatchSnapshot()
    })

    test("doesn't render the permissions alert on chrome://newtab", () => {
        expect(
            renderer.create(
                <OptionsMenu
                    {...stubs}
                    currentTabStatus={{ host: 'newtab', protocol: 'chrome:', hasPermissions: false }}
                />
            )
        ).toMatchSnapshot()
    })

    test("doesn't render the permissions alert on about://addons", () => {
        expect(
            renderer.create(
                <OptionsMenu
                    {...stubs}
                    currentTabStatus={{ host: 'addons', protocol: 'about:', hasPermissions: false }}
                />
            )
        ).toMatchSnapshot()
    })

    test('fires requestPermissions', () => {
        const requestPermissions = sinon.spy()
        const { container } = render(
            <OptionsMenu
                {...stubs}
                currentTabStatus={{ host: 'gitlab.com', protocol: 'http', hasPermissions: false }}
                requestPermissions={requestPermissions}
            />
        )
        const requestLink = container.querySelector('.request-permissions__test')!
        fireEvent.click(requestLink)
        expect(requestPermissions.calledOnce).toBe(true)
    })

    test('renders the feature flags', () => {
        expect(
            renderer.create(
                <OptionsMenu
                    {...stubs}
                    isSettingsOpen={true}
                    featureFlags={[
                        { key: 'foo', value: true },
                        { key: 'bar', value: false },
                    ]}
                />
            )
        ).toMatchSnapshot()
    })

    test('triggers the toggleFeatureFlag handler', () => {
        const toggleFeatureFlag = sinon.spy()
        const { container } = render(
            <OptionsMenu
                {...stubs}
                isSettingsOpen={true}
                featureFlags={[
                    { key: 'foo', value: true },
                    { key: 'bar', value: false },
                ]}
                toggleFeatureFlag={toggleFeatureFlag}
            />
        )
        const fooCheckbox = container.querySelector('#foo')!
        fireEvent.click(fooCheckbox)
        expect(toggleFeatureFlag.calledOnce).toBe(true)
    })
})
