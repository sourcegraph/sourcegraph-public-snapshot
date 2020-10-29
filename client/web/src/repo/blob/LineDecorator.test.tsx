import React from 'react'
import ReactDOM from 'react-dom'
import renderer from 'react-test-renderer'
import { TextDocumentDecoration, ThemableDecorationStyle } from 'sourcegraph'
import { Position, Range } from '@sourcegraph/extension-api-classes'
import { LineDecorator, LineDecoratorProps } from './LineDecorator'
import sinon from 'sinon'

describe('LineDecorator', () => {
    beforeAll(() => {
        // react-test-renderer doesn't support portals (https://github.com/facebook/react/issues/11565),
        // so just pass children through
        ReactDOM.createPortal = sinon.spy(children => children)
    })

    function createCodeElement() {
        const codeElement = document.createElement('td')
        codeElement.classList.add('code')
        const parentRow = document.createElement('tr')
        parentRow.append(codeElement)
        return { codeElement, parentRow }
    }

    function createLineDecoratorProps(
        line: number,
        decorations: TextDocumentDecoration[],
        /** An HTMLTableCellElement that must be a child of an HTMLTableRowElement */
        codeElement = createCodeElement().codeElement
    ): LineDecoratorProps {
        const codeViewElement = document.createElement('div')

        return {
            line,
            decorations,
            portalID: `line-decoration-attachment-${line}`,
            isLightTheme: false,
            codeViewReference: { current: codeViewElement },
            getCodeElementFromLineNumber: (codeView: HTMLElement, lineNumber: number): HTMLTableCellElement | null => {
                if (codeView === codeViewElement && lineNumber === line) {
                    return codeElement
                }

                return null
            },
        }
    }

    it('renders one attachment', () => {
        const props = createLineDecoratorProps(1, [
            { after: { contentText: 'test content' }, range: new Range(new Position(0, 0), new Position(0, 0)) },
        ])

        expect(renderer.create(<LineDecorator {...props} />).toJSON()).toMatchSnapshot()
    })

    it('renders multiple attachments', () => {
        const props = createLineDecoratorProps(1, [
            {
                after: { contentText: 'attachment from extension one' },
                range: new Range(new Position(0, 0), new Position(0, 0)),
            },
            {
                after: { contentText: 'attachment from extension two' },
                range: new Range(new Position(0, 0), new Position(0, 0)),
            },
        ])

        expect(renderer.create(<LineDecorator {...props} />).toJSON()).toMatchSnapshot()
    })

    it('decorates line', () => {
        const { codeElement, parentRow } = createCodeElement()
        const themeableDecorationStyle: ThemableDecorationStyle = {
            backgroundColor: 'black',
            borderColor: 'teal',
        }

        const props = createLineDecoratorProps(
            1,
            [
                {
                    range: new Range(new Position(0, 0), new Position(0, 0)),
                    ...themeableDecorationStyle,
                },
            ],
            codeElement
        )

        const testRenderer = renderer.create(<LineDecorator {...props} />)

        // Code row should be styled after the decorator mounts
        expect({
            backgroundColor: parentRow.style.backgroundColor,
            borderColor: parentRow.style.borderColor,
        }).toStrictEqual(themeableDecorationStyle)

        // Code row should be unstyled after the decorator unmounts
        testRenderer.unmount()
        expect({
            backgroundColor: parentRow.style.backgroundColor,
            borderColor: parentRow.style.borderColor,
        }).toStrictEqual({ backgroundColor: '', borderColor: '' })
    })

    it('updates decorations on theme change', () => {
        const { codeElement, parentRow } = createCodeElement()
        const themeableDecorationStyleLight: ThemableDecorationStyle = {
            backgroundColor: 'black',
            borderColor: 'teal',
        }
        const themeableDecorationStyleDark: ThemableDecorationStyle = {
            backgroundColor: 'white',
            borderColor: 'pink',
        }

        const props = createLineDecoratorProps(
            1,
            [
                {
                    range: new Range(new Position(0, 0), new Position(0, 0)),
                    light: themeableDecorationStyleLight,
                    dark: themeableDecorationStyleDark,
                },
            ],
            codeElement
        )

        const testRenderer = renderer.create(<LineDecorator {...props} isLightTheme={true} />)

        expect({
            backgroundColor: parentRow.style.backgroundColor,
            borderColor: parentRow.style.borderColor,
        }).toStrictEqual(themeableDecorationStyleLight)

        testRenderer.update(<LineDecorator {...props} isLightTheme={false} />)

        expect({
            backgroundColor: parentRow.style.backgroundColor,
            borderColor: parentRow.style.borderColor,
        }).toStrictEqual(themeableDecorationStyleDark)
    })
})
