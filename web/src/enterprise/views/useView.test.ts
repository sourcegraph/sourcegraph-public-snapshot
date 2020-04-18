import { renderHook } from '@testing-library/react-hooks'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { useView } from './useView'
import { of } from 'rxjs'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { View } from '../../../../shared/src/api/client/services/viewService'

const VIEW: View = { title: 't', content: [{ kind: MarkupKind.Markdown, value: 'c' }] }
const EMPTY_PARAMS: Parameters<typeof useView>[2] = {}

describe('useView', () => {
    test('returns view', () => {
        const contributions: Parameters<typeof useView>[3] = of({
            views: [{ id: 'v', where: ContributableViewContainer.Global }],
        })
        const viewService: Parameters<typeof useView>[4] = {
            get: () => of(VIEW),
        }
        const { result } = renderHook(() =>
            useView('v', ContributableViewContainer.Global, EMPTY_PARAMS, contributions, viewService)
        )
        expect(result.current).toBe(VIEW)
    })

    test('reports view is loading', () => {
        const contributions: Parameters<typeof useView>[3] = of({
            views: [],
        })
        const viewService: Parameters<typeof useView>[4] = {
            get: () => of(null),
        }
        const { result } = renderHook(() =>
            useView('v', ContributableViewContainer.Global, EMPTY_PARAMS, contributions, viewService)
        )
        expect(result.current).toBe(undefined)
    })
})
