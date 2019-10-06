import React from 'react'
import { RuleActionType, RuleActionTypeComponentContext } from '..'
import PencilBoxIcon from 'mdi-react/PencilBoxIcon'

interface EditorRuleAction {
    type: 'editor'
}

const EditorRuleActionFormControl: React.FunctionComponent<RuleActionTypeComponentContext<EditorRuleAction>> = () => (
    <div className="form-group">
        <label>Show campaign diagnostics in editors</label>
        <small className="form-help text-muted">
            <a
                href="https://docs.sourcegraph.com/integration/editor"
                rel="noopener noreferrer"
                target="_blank"
                className="btn btn-primary"
            >
                Install Sourcegraph editor integrations
            </a>
        </small>
    </div>
)

export const EditorRuleActionType: RuleActionType<'editor', EditorRuleAction> = {
    id: 'editor',
    title: 'Editor',
    icon: PencilBoxIcon,
    renderForm: EditorRuleActionFormControl,
    initialValue: {
        type: 'editor',
    },
}
