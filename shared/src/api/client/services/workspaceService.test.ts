import { createWorkspaceService } from './workspaceService'

describe('WorkspaceService', () => {
    test('roots', () => {
        const workspaceService = createWorkspaceService()
        expect(workspaceService.roots.value).toEqual([])

        workspaceService.roots.next([{ uri: 'a', baseUri: 'b' }])
        expect(workspaceService.roots.value).toEqual([{ uri: 'a', baseUri: 'b' }])
    })
})
