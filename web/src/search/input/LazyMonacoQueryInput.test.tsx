import React from 'react'
import { noop } from 'lodash'
import { PlainQueryInput } from './LazyMonacoQueryInput'
import { createMemoryHistory } from 'history'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'

describe('PlainQueryInput', () => {
    const history = createMemoryHistory()
    test('empty', () =>
        expect(
            mount(
                <PlainQueryInput
                    history={history}
                    location={history.location}
                    queryState={{
                        query: '',
                        cursorPosition: 0,
                    }}
                    patternType={SearchPatternType.regexp}
                    setPatternType={noop}
                    caseSensitive={true}
                    setCaseSensitivity={noop}
                    onChange={noop}
                    onSubmit={noop}
                    isLightTheme={false}
                    settingsCascade={{ subjects: [], final: {} }}
                    copyQueryButton={false}
                    versionContext={undefined}
                />
            )
        ).toMatchSnapshot())

    test('with query', () =>
        expect(
            mount(
                <PlainQueryInput
                    history={history}
                    location={history.location}
                    queryState={{
                        query: 'repo:jsonrpc2 file:async.go asyncHandler',
                        cursorPosition: 0,
                    }}
                    patternType={SearchPatternType.regexp}
                    setPatternType={noop}
                    caseSensitive={true}
                    setCaseSensitivity={noop}
                    onChange={noop}
                    onSubmit={noop}
                    isLightTheme={false}
                    settingsCascade={{ subjects: [], final: {} }}
                    copyQueryButton={false}
                    versionContext={undefined}
                />
            )
        ).toMatchSnapshot())
})
