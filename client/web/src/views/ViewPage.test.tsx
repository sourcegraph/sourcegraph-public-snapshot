import React from 'react'
import { create, act } from 'react-test-renderer'
import { ViewPage } from './ViewPage'
import * as H from 'history'
import { Controller } from '../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../shared/src/graphql-operations'
import { noop } from 'lodash'

jest.mock('@sourcegraph/react-loading-spinner', () => ({ LoadingSpinner: 'LoadingSpinner' }))
jest.mock('./QueryInputInViewContent', () => ({ QueryInputInViewContent: 'QueryInputInViewContent' }))

const commonProps: Omit<React.ComponentProps<typeof ViewPage>, 'viewID' | 'extraPath' | '_getView'> = {
    settingsCascade: { final: {}, subjects: [] },
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    setCaseSensitivity: () => undefined,
    setPatternType: () => undefined,
    history: H.createMemoryHistory(),
    location: H.createLocation('/'),
    extensionsController: { services: { contribution: { getContributions: () => ({}) } } } as Controller,
    telemetryService: { log: noop, logViewEvent: noop },
    copyQueryButton: false,
    versionContext: undefined,
    selectedSearchContextSpec: 'global',
    globbing: false,
}

// TODO: Update these tests
describe.skip('ViewPage', () => {
    test('view is loading', () => {
        const renderer = create(<ViewPage {...commonProps} viewID="v" extraPath="" )
        act(() => undefined)
        expect(renderer.toJSON()).toMatchSnapshot()
    })

    test('view not found', () => {
        const renderer = create(<ViewPage {...commonProps} viewID="v" extraPath="" />)
        act(() => undefined)
        expect(renderer.toJSON()).toMatchSnapshot()
    })

    test('renders view', () => {
        // const mockView = {
        //     title: 't',
        //     content: [
        //         {
        //             kind: MarkupKind.Markdown,
        //             value: '**a**',
        //         },
        //         {
        //             kind: MarkupKind.PlainText,
        //             value: '*b*',
        //         },
        //         {
        //             value: '*c*',
        //         },
        //         {
        //             component: 'QueryInput',
        //             props: { implicitQueryPrefix: 'x' },
        //         },
        //     ],
        // }

        const renderer = create(
            <ViewPage
                {...commonProps}
                viewID="v"
                extraPath=""
            />
        )
        act(() => undefined)
        expect(renderer.toJSON()).toMatchSnapshot()
    })
})
