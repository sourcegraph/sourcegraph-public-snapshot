import * as vscode from 'vscode'

import { Diff } from './diff'
import { FixupFile } from './FixupFile'
import { CodyTaskState } from './utils'

export type taskID = string

export class FixupTask {
    public id: taskID
    public state_: CodyTaskState = CodyTaskState.idle
    // The original text that we're working on updating. Set when we start an
    // LLM spin.
    public original = ''
    // The text of the streaming turn of the LLM, if any
    public inProgressReplacement: string | undefined
    // The text of the last completed turn of the LLM, if any
    public replacement: string | undefined
    // If text has been received from the LLM and a diff has been computed, it
    // is cached here. Diffs are recomputed lazily and may be stale.
    public diff: Diff | undefined
    // The number of times we've submitted this to the LLM.
    public spinCount = 0

    constructor(
        public readonly fixupFile: FixupFile,
        public readonly instruction: string,
        public selectionRange: vscode.Range
    ) {
        this.id = Date.now().toString(36).replace(/\d+/g, '')
    }

    /**
     * Sets the task state. Checks the state transition is valid.
     */
    public set state(state: CodyTaskState) {
        if (this.state_ === CodyTaskState.error) {
            throw new Error('invalid transition out of error sink state')
        }
        this.state_ = state
    }

    public get state(): CodyTaskState {
        return this.state_
    }
}
