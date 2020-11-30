import * as sourcegraph from 'sourcegraph'
import { WorkbenchViewScheduler } from './scheduler'

export abstract class WorkbenchView<TState extends object> {
    abstract state: TState

    constructor(private scheduler: WorkbenchViewScheduler, private viewType: sourcegraph.WorkbenchViewType) {}

    public update(nextState: Partial<TState>): void {
        this.state = { ...this.state, ...nextState }
        this.scheduler.schedule({ viewType: this.viewType, type: 'update' })
    }
}

export class ExtensionStatusBarItem
    extends WorkbenchView<sourcegraph.StatusBarItemState>
    implements sourcegraph.StatusBarItem {
    public state: sourcegraph.StatusBarItemState = {
        visible: true,
        contentText: '',
        hoverMessage: '',
        linkURL: '',
    }

    constructor(scheduler: WorkbenchViewScheduler, public unsubscribe: () => void) {
        super(scheduler, 'statusBarItem')
    }
}
