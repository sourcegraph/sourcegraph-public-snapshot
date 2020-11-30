import { noop } from 'lodash'
import { WorkbenchViewType } from 'sourcegraph'
import { createWorkbenchViewScheduler, WorkbenchViewScheduler } from './scheduler'
import { WorkbenchView } from './views'

// Create a minimal workbench view that extends the base class in order to test
// shared behavior of workbench views and to avoid duplicating tests.
// Any functionality specific to a workbench view type can be tested separately
// or in an integration test.

interface MockWorkbenchViewState {
    contentText: string
    visible: boolean
}

class MockWorkbenchView extends WorkbenchView<MockWorkbenchViewState> {
    public state: MockWorkbenchViewState = {
        contentText: '',
        visible: true,
    }

    constructor(scheduler: WorkbenchViewScheduler, public unsubscribe: () => void) {
        super(scheduler, 'mockWorkbenchViewType' as WorkbenchViewType)
    }
}

describe('Workbench View', () => {
    const scheduler = createWorkbenchViewScheduler(noop, () => undefined)

    afterAll(() => {
        scheduler.unsubscribe()
    })

    test('updating state', () => {
        const viewInstance = new MockWorkbenchView(scheduler, noop)

        expect(viewInstance.state).toStrictEqual({ contentText: '', visible: true })

        // partial update
        viewInstance.update({ contentText: 'partial' })
        expect(viewInstance.state).toStrictEqual({ contentText: 'partial', visible: true })

        // full update
        viewInstance.update({ contentText: 'hidden', visible: false })
        expect(viewInstance.state).toStrictEqual({ contentText: 'hidden', visible: false })
    })
})
