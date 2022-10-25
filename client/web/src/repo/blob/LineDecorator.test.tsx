import { render, within, waitFor } from '@testing-library/react'
import { ReplaySubject } from 'rxjs'
import { TextDocumentDecoration, ThemableDecorationStyle } from 'sourcegraph'

import { Position, Range } from '@sourcegraph/extension-api-classes'

import { LineDecorator, LineDecoratorProps } from './LineDecorator'

describe('LineDecorator', () => {
    function createCodeElement() {
        const parentRow = document.createElement('tr')

        const lineElement = parentRow.insertCell()
        lineElement.classList.add('line')
        lineElement.dataset.line = '1'

        const codeElement = parentRow.insertCell()
        codeElement.classList.add('code')

        return { parentRow, lineElement, codeElement }
    }

    function createLineDecoratorProps(
        line: number,
        decorations: TextDocumentDecoration[],
        /** An HTMLTableCellElement that must be a child of an HTMLTableRowElement */
        codeElement = createCodeElement().codeElement
    ): LineDecoratorProps {
        const codeViewElement = document.createElement('div')

        const codeViewElements = new ReplaySubject<HTMLElement | null>()
        codeViewElements.next(codeViewElement)

        return {
            line,
            decorations,
            portalID: `line-decoration-attachment-${line}`,
            isLightTheme: false,
            codeViewElements,
            getCodeElementFromLineNumber: (codeView: HTMLElement, lineNumber: number): HTMLTableCellElement | null => {
                if (codeView === codeViewElement && lineNumber === line) {
                    return codeElement
                }

                return null
            },
        }
    }

    it('renders one attachment', () => {
        const { codeElement } = createCodeElement()
        const props = createLineDecoratorProps(
            1,
            [{ after: { contentText: 'test content' }, range: new Range(new Position(0, 0), new Position(0, 0)) }],
            codeElement
        )

        render(<LineDecorator {...props} />)

        const container = within(codeElement)
        waitFor(() => expect(container.getByTestId('line-decoration')).toBeVisible())
        expect(codeElement).toMatchSnapshot()
    })

    it('renders multiple attachments', () => {
        const { codeElement } = createCodeElement()
        const props = createLineDecoratorProps(
            1,
            [
                {
                    after: { contentText: 'attachment from extension one' },
                    range: new Range(new Position(0, 0), new Position(0, 0)),
                },
                {
                    after: { contentText: 'attachment from extension two' },
                    range: new Range(new Position(0, 0), new Position(0, 0)),
                },
            ],
            codeElement
        )

        render(<LineDecorator {...props} />)

        const container = within(codeElement)
        waitFor(() => expect(container.getByTestId('line-decoration')).toBeVisible())
        expect(codeElement).toMatchSnapshot()
    })

    it('decorates line', () => {
        const { codeElement } = createCodeElement()
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

        const wrapper = render(<LineDecorator {...props} />)

        // Code element should be styled after the decorator mounts
        expect({
            backgroundColor: codeElement.style.backgroundColor,
            borderColor: codeElement.style.borderColor,
        }).toStrictEqual(themeableDecorationStyle)

        // Code element should be unstyled after the decorator unmounts
        wrapper.unmount()
        expect({
            backgroundColor: codeElement.style.backgroundColor,
            borderColor: codeElement.style.borderColor,
        }).toStrictEqual({ backgroundColor: '', borderColor: '' })
    })

    it('updates decorations on theme change', () => {
        const { codeElement } = createCodeElement()
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

        const { rerender } = render(<LineDecorator {...props} isLightTheme={true} />)

        expect({
            backgroundColor: codeElement.style.backgroundColor,
            borderColor: codeElement.style.borderColor,
        }).toStrictEqual(themeableDecorationStyleLight)

        rerender(<LineDecorator {...props} isLightTheme={false} />)

        expect({
            backgroundColor: codeElement.style.backgroundColor,
            borderColor: codeElement.style.borderColor,
        }).toStrictEqual(themeableDecorationStyleDark)
    })
})
