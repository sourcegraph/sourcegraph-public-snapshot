import * as sourcegraph from 'sourcegraph'

export interface ChangesetPlan {
    // TODO!(sqs): always assume only 1 operation
    operations: ChangesetPlanOperation[]
}

export interface ChangesetPlanOperation {
    /**
     * A message that describes what this operation does (similar to a commit message).
     */
    message: string

    /**
     * A query that describes a set of diagnostics that this operation resolves. The action's
     * command is executed with the matching diagnostics as arguments.
     */
    diagnostics?: sourcegraph.DiagnosticQuery

    /**
     * A command that, when executed, returns a {@link sourcegraph.WorkspaceEdit}. The edits from a
     * changeset's operations are applied sequentially to determine the diff.
     *
     * If a {@link ChangesetPlanOperation#diagnosticQuery} is specified, the edits are assumed to
     * resolve all diagnostics that match the query.
     */
    editCommand: Pick<sourcegraph.Command, 'command' | 'arguments'>
}
