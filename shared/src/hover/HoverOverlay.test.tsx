import { MarkupKind } from '@sourcegraph/extension-api-classes'
import * as H from 'history'
import React from 'react'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { HoverOverlay, HoverOverlayProps } from './HoverOverlay'
import { NEVER } from 'rxjs'
import { subtypeOf } from '../util/types'
import { mount, shallow } from 'enzyme'

describe('HoverOverlay', () => {
    const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve() }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined, settings: NEVER }
    const history = H.createMemoryHistory({ keyLength: 0 })
    const commonProps = subtypeOf<HoverOverlayProps<string>>()({
        location: history.location,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        extensionsController: NOOP_EXTENSIONS_CONTROLLER,
        platformContext: NOOP_PLATFORM_CONTEXT,
        showCloseButton: false,
        hoveredToken: { repoName: 'r', commitID: 'c', revision: 'v', filePath: 'f', line: 1, character: 2 },
        overlayPosition: { left: 0, top: 0 },
        isLightTheme: false,
    })

    test('actions and hover undefined', () => {
        expect(mount(<HoverOverlay {...commonProps} />)).toMatchSnapshot()
    })

    test('actions loading', () => {
        expect(mount(<HoverOverlay {...commonProps} actionsOrError="loading" />)).toMatchSnapshot()
    })

    test('hover loading', () => {
        expect(mount(<HoverOverlay {...commonProps} hoverOrError="loading" />)).toMatchSnapshot()
    })

    test('actions and hover loading', () => {
        expect(
            mount(<HoverOverlay {...commonProps} actionsOrError="loading" hoverOrError="loading" />)
        ).toMatchSnapshot()
    })

    test('actions empty', () => {
        const component = mount(<HoverOverlay {...commonProps} actionsOrError={[]} />)
        expect(component).toMatchSnapshot()
    })

    test('hover empty', () => {
        expect(mount(<HoverOverlay {...commonProps} hoverOrError={null} />)).toMatchSnapshot()
    })

    test('actions and hover empty', () => {
        expect(mount(<HoverOverlay {...commonProps} actionsOrError={[]} hoverOrError={null} />)).toMatchSnapshot()
    })

    test('actions present', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={[{ action: { id: 'a', command: 'c', title: 'Some title' } }]}
                />
            )
        ).toMatchSnapshot()
    })

    test('hover present', () => {
        expect(
            mount(
                <HoverOverlay
                    {...commonProps}
                    hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                />
            )
        ).toMatchSnapshot()
    })

    test('multiple hovers present', () => {
        expect(
            mount(
                <HoverOverlay
                    {...commonProps}
                    hoverOrError={{
                        contents: [
                            { kind: MarkupKind.Markdown, value: 'v' },
                            { kind: MarkupKind.Markdown, value: 'v2' },
                        ],
                    }}
                />
            )
        ).toMatchSnapshot()
    })

    test('actions and hover present', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={[{ action: { id: 'a', command: 'c' } }]}
                    hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                />
            )
        ).toMatchSnapshot()
    })

    test('actions, hover and alert present', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={[{ action: { id: 'a', command: 'c' } }]}
                    hoverOrError={{
                        contents: [{ kind: MarkupKind.Markdown, value: 'v' }],
                        alerts: [
                            {
                                type: 'a' as const,
                                content: (
                                    <>
                                        b <small>c</small> <code>d</code>
                                    </>
                                ),
                            },
                        ],
                    }}
                />
            )
        ).toMatchSnapshot()
    })

    test('actions present, hover loading', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={[{ action: { id: 'a', command: 'c' } }]}
                    hoverOrError="loading"
                />
            )
        ).toMatchSnapshot()
    })

    test('hover present, actions loading', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError="loading"
                    hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                />
            )
        ).toMatchSnapshot()
    })

    test('actions error', () => {
        expect(
            shallow(<HoverOverlay {...commonProps} actionsOrError={{ message: 'm', name: 'c' }} />)
        ).toMatchSnapshot()
    })

    test('hover error', () => {
        expect(shallow(<HoverOverlay {...commonProps} hoverOrError={{ message: 'm', name: 'c' }} />)).toMatchSnapshot()
    })

    test('actions and hover error', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={{ message: 'm1', name: 'c1' }}
                    hoverOrError={{ message: 'm2', name: 'c2' }}
                />
            )
        ).toMatchSnapshot()
    })

    test('actions error, hover present', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={{ message: 'm', name: 'c' }}
                    hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                />
            )
        ).toMatchSnapshot()
    })

    test('hover error, actions present', () => {
        expect(
            shallow(
                <HoverOverlay
                    {...commonProps}
                    actionsOrError={[{ action: { id: 'a', command: 'c' } }]}
                    hoverOrError={{ message: 'm', name: 'c' }}
                />
            )
        ).toMatchSnapshot()
    })
})
