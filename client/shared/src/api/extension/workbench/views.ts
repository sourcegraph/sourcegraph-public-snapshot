import * as sourcegraph from 'sourcegraph'
import { WorkbenchViewScheduler } from './scheduler'

abstract class WorkbenchView<TState extends object> {
    abstract state: TState

    constructor(private scheduler: WorkbenchViewScheduler, private viewType: sourcegraph.WorkbenchViewType) {}

    public update(nextState: Partial<TState>): void {
        this.state = { ...this.state, ...nextState }
        this.scheduler.schedule({ viewType: this.viewType, type: 'update' })
    }
}

// FOR NOW finish naive impl (cloning all view state)
// TODO(tj): Where should we keep migrated services?

// VIEWS

// Why am I using classes? Saves memory (prototype methods)

// Implement merging state setter so extensions can batch-update
// interface StatusBarItemState extends Pick<sourcegraph.StatusBarItem, 'contentText' | 'hoverMessage' | 'linkURL'> {}

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

    // TODO(tj): don't need these getters and setters
    // Will increase development speed for views with complex state

    // public get visible(): boolean {
    //     return this.state.visible
    // }

    // public set visible(visible: boolean) {
    //     this.update({ visible })
    // }

    // public get contentText(): string {
    //     return this.state.contentText
    // }

    // public set contentText(contentText: string) {
    //     this.update({ contentText })
    // }

    // public get hoverMessage(): string | undefined {
    //     return this.state.hoverMessage
    // }

    // public set hoverMessage(hoverMessage: string | undefined) {
    //     this.update({ hoverMessage })
    // }

    // public get linkURL(): string | undefined {
    //     return this.state.linkURL
    // }

    // public set linkURL(linkURL: string | undefined) {
    //     this.update({ linkURL })
    // }
}
