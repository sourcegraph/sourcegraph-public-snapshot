import { from } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'
import { isDefined } from '../../util/types'
import { ViewComponentData } from '../client/model'
import { assertToJSON } from '../extension/types/testHelpers'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

const withSelections = (...selections: { start: number; end: number }[]): ViewComponentData => ({
    type: 'textEditor',
    item: { uri: 'foo', languageId: 'l1', text: 't1' },
    selections: selections.map(({ start, end }) => ({
        start: {
            line: start,
            character: 0,
        },
        end: {
            line: end,
            character: 0,
        },
        anchor: {
            line: start,
            character: 0,
        },
        active: {
            line: end,
            character: 0,
        },
        isReversed: false,
    })),
    isActive: true,
})

describe('Selections (integration)', () => {
    describe('editor.selectionsChanged', () => {
        test('reflects changes to the current selections', async () => {
            const { model, extensionHost } = await integrationTestContext(undefined, {
                roots: [],
                visibleViewComponents: [],
            })
            const selectionChanges = from(extensionHost.app.activeWindowChanges).pipe(
                filter(isDefined),
                switchMap(window => window.activeViewComponentChanges),
                filter(isDefined),
                switchMap(editor => editor.selectionsChanges)
            )
            const selectionValues = collectSubscribableValues(selectionChanges)
            const testValues = [
                [{ start: 3, end: 5 }],
                [{ start: 1, end: 10 }, { start: 25, end: 40 }, { start: 56, end: 57 }],
                [],
            ]
            for (const selections of testValues) {
                model.next({
                    ...model.value,
                    visibleViewComponents: [withSelections(...selections)],
                })
                await extensionHost.internal.sync()
            }
            assertToJSON(
                selectionValues.map(selections => selections.map(s => ({ start: s.start.line, end: s.end.line }))),
                testValues
            )
        })
    })
})
