import React from 'react'
import * as H from 'history'
import { Toggles } from './Toggles'
import { mount } from 'enzyme'
import { SearchPatternType } from '../../../graphql-operations'

describe('Query input toggle state', () => {
    test('case toggle for case subexpressions', () => {
        expect(
            mount(
                <Toggles
                    navbarSearchQuery="(case:yes foo) or (case:no bar)"
                    history={H.createBrowserHistory()}
                    location={H.createBrowserHistory().location}
                    patternType={SearchPatternType.literal}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    settingsCascade={{ subjects: null, final: {} }}
                    copyQueryButton={false}
                    versionContext={undefined}
                />
            ).find('.test-case-sensitivity-toggle')
        ).toMatchSnapshot()
    })

    test('case toggle for patterntype subexpressions', () => {
        expect(
            mount(
                <Toggles
                    navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                    history={H.createBrowserHistory()}
                    location={H.createBrowserHistory().location}
                    patternType={SearchPatternType.literal}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    settingsCascade={{ subjects: null, final: {} }}
                    copyQueryButton={false}
                    versionContext={undefined}
                />
            ).find('.test-case-sensitivity-toggle')
        ).toMatchSnapshot()
    })

    test('regexp toggle for patterntype subexpressions', () => {
        expect(
            mount(
                <Toggles
                    navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                    history={H.createBrowserHistory()}
                    location={H.createBrowserHistory().location}
                    patternType={SearchPatternType.literal}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    settingsCascade={{ subjects: null, final: {} }}
                    copyQueryButton={false}
                    versionContext={undefined}
                />
            ).find('.test-regexp-toggle')
        ).toMatchSnapshot()
    })
})
