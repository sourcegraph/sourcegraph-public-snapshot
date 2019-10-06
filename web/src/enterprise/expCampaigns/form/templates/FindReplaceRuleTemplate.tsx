import FindReplaceIcon from 'mdi-react/FindReplaceIcon'
import { FindReplaceCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/findReplace'
import React, { useCallback, useEffect } from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'
import { RuleDefinition } from '../../../rulesOLD/types'
import TextareaAutosize from 'react-textarea-autosize'

const TEMPLATE_ID = 'findReplace'

interface Props extends RuleTemplateComponentContext {}

const FindReplaceCampaignTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    onCampaignChange,
    disabled,
    location,
}) => {
    const context: FindReplaceCampaignContext | undefined = value.template ? value.template.context : undefined

    const updateContext = useCallback(
        (update: Partial<FindReplaceCampaignContext>): void => {
            const newContext = { ...context, ...update }
            onChange({
                name: 'Find-replace',
                template: {
                    template: TEMPLATE_ID,
                    context: newContext,
                },
                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                definition: JSON.stringify({
                    type: 'ActionRule',
                    context: newContext,
                    action: 'findReplace.rewrite',
                } as RuleDefinition),
            })
            onCampaignChange({
                isValid: newContext.matchTemplate !== undefined && newContext.rewriteTemplate !== undefined,
            })
        },
        [context, onCampaignChange, onChange]
    )

    // Set defaults.
    const locationSearch = location ? location.search : ''
    useEffect(() => {
        if (context === undefined) {
            const update: Partial<FindReplaceCampaignContext> = {}

            const params = new URLSearchParams(locationSearch)
            const matchTemplate = params.get('matchTemplate')
            if (matchTemplate !== null) {
                update.matchTemplate = matchTemplate
            }
            const rule = params.get('rule')
            if (rule !== null) {
                update.rule = rule
            }
            const rewriteTemplate = params.get('rewriteTemplate')
            if (rewriteTemplate !== null) {
                update.rewriteTemplate = rewriteTemplate
            }
            updateContext(update)

            if (value.name === '') {
                onCampaignChange({ name: 'Find-replace' })
            }
        }
    }, [context, locationSearch, onCampaignChange, onChange, updateContext, value.name])

    const onMatchTemplateChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => updateContext({ matchTemplate: e.currentTarget.value }),
        [updateContext]
    )

    const onRuleChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => updateContext({ rule: e.currentTarget.value }),
        [updateContext]
    )

    const onRewriteTemplateChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => updateContext({ rewriteTemplate: e.currentTarget.value }),
        [updateContext]
    )

    if (context === undefined) {
        return null
    }

    return (
        <>
            <div className="form-group">
                <label htmlFor="campaign-template-form__matchTemplate">Match template</label>
                <TextareaAutosize
                    id="campaign-template-form__matchTemplate"
                    className="form-control"
                    required={true}
                    minLength={1}
                    minRows={3}
                    value={context.matchTemplate || ''}
                    onChange={onMatchTemplateChange}
                    autoFocus={true}
                    disabled={disabled}
                />
                <p className="form-help text-muted small mb-0">
                    <a href="https://comby.dev/#match-syntax" target="_blank" rel="noopener noreferrer">
                        Comby syntax supported
                    </a>
                </p>
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__rule">Rule</label>
                <TextareaAutosize
                    type="text"
                    id="campaign-template-form__rule"
                    className="form-control"
                    value={context.rule || ''}
                    onChange={onRuleChange}
                    disabled={disabled}
                />
                <p className="form-help text-muted small mb-0">
                    <a href="https://comby.dev/#advanced-usage" target="_blank" rel="noopener noreferrer">
                        Comby rules supported
                    </a>
                </p>
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__rewriteTemplate">Rewrite template</label>
                <TextareaAutosize
                    type="text"
                    id="campaign-template-form__rewriteTemplate"
                    className="form-control"
                    minRows={3}
                    value={context.rewriteTemplate || ''}
                    onChange={onRewriteTemplateChange}
                    disabled={disabled}
                />
                <p className="form-help text-muted small mb-0">
                    <a href="https://comby.dev/#match-syntax" target="_blank" rel="noopener noreferrer">
                        Comby syntax supported
                    </a>
                </p>
            </div>
        </>
    )
}

export const FindReplaceRuleTemplate: RuleTemplate = {
    id: 'findReplace',
    title: 'Find-replace',
    detail:
        'Configurable find-replace across multiple files and repositories, opening changesets with the diffs. [Comby syntax](https://comby.dev/#match-syntax) supported for matching and rewriting.',
    icon: FindReplaceIcon,
    renderForm: FindReplaceCampaignTemplateForm,
}
