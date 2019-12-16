import React from 'react'
import { CampaignType } from './backend'
import * as jsonc from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { ErrorAlert } from '../../../components/alerts'

const defaultFormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

interface CampaignFormProps {
    campaignArguments: string
    // TODO use generated types
    parsed: any
    onChange: (newText: string) => void
}

const CombyForm: React.FunctionComponent<CampaignFormProps> = ({ campaignArguments, parsed, onChange }) => {
    // TODO use keyof generated types
    const createChangeHandler = (prop: string) => (event: React.ChangeEvent<{ value: string }>) =>
        onChange(
            jsonc.applyEdits(
                campaignArguments,
                setProperty(campaignArguments, [prop], event.target.value, defaultFormattingOptions)
            )
        )

    return (
        <>
            <div className="form-group row">
                <label className="col-sm-2 col-form-label">Scope query</label>
                <div className="col-sm-10">
                    <input
                        type="text"
                        className="form-control"
                        value={parsed.scopeQuery}
                        onChange={createChangeHandler('scopeQuery')}
                    />
                </div>
            </div>
            <div className="form-group row">
                <label className="col-sm-2 col-form-label">Match template</label>
                <div className="col-sm-10">
                    <textarea
                        className="form-control"
                        value={parsed.matchTemplate}
                        onChange={createChangeHandler('matchTemplate')}
                        required={true}
                    />
                    <small className="text-muted">
                        <a rel="noopener noreferrer" target="_blank" href="https://comby.dev/#match-syntax">
                            Learn about comby syntax
                        </a>
                    </small>
                </div>
            </div>
            <div className="form-group row">
                <label className="col-sm-2 col-form-label">Rewrite template</label>
                <div className="col-sm-10">
                    <textarea
                        className="form-control"
                        value={parsed.rewriteTemplate}
                        onChange={createChangeHandler('rewriteTemplate')}
                    />
                    <small className="text-muted">
                        <a rel="noopener noreferrer" target="_blank" href="https://comby.dev/#match-syntax">
                            Learn about comby syntax
                        </a>
                    </small>
                </div>
            </div>
        </>
    )
}

const CredentialsForm: React.FunctionComponent<CampaignFormProps> = ({ campaignArguments, parsed, onChange }) => {
    if (!parsed.matchers) {
        try {
            onChange(
                jsonc.applyEdits(
                    campaignArguments,
                    setProperty(campaignArguments, ['matchers'], [{ type: 'npm' }], defaultFormattingOptions)
                )
            )
        } catch {
            return <ErrorAlert error="Invalid JSON. Please use JSON editor to fix." />
        }
        return null
    }
    if (!Array.isArray(parsed.matchers)) {
        return <ErrorAlert error="Invalid JSON: matchers is not an array. Please use JSON editor to fix." />
    }
    if (parsed.matchers.length === 0) {
        try {
            onChange(
                jsonc.applyEdits(
                    campaignArguments,
                    setProperty(campaignArguments, ['matchers', 0], { type: 'npm' }, defaultFormattingOptions)
                )
            )
        } catch {
            return <ErrorAlert error="Invalid JSON. Please use JSON editor to fix." />
        }
        return null
    }
    if (parsed.matchers.length > 1) {
        console.log(parsed)
        return <ErrorAlert error="Invalid JSON: more than one matcher. Please use JSON editor to fix." />
    }
    return (
        <>
            {/* <div className="form-group row">
                <label className="col-sm-2 col-form-label">Type</label>
                <div className="col-sm-10">
                    <select className="form-control" value="npm">
                        <option value="npm">NPM</option>
                    </select>
                </div>
            </div> */}
            {parsed.matchers.map((matcher: any, i: number) => (
                <div className="form-group row" key={i}>
                    <label className="col-sm-2 col-form-label">Replace with</label>
                    <div className="col-sm-10">
                        <input
                            type="text"
                            className="form-control"
                            value={parsed.matchers[i].replaceWith}
                            onChange={event =>
                                onChange(
                                    jsonc.applyEdits(
                                        campaignArguments,
                                        setProperty(
                                            campaignArguments,
                                            ['matchers', i, 'replaceWith'],
                                            event.target.value,
                                            defaultFormattingOptions
                                        )
                                    )
                                )
                            }
                        />
                    </div>
                </div>
            ))}
        </>
    )
}

const components: Record<CampaignType, React.ComponentType<CampaignFormProps>> = {
    comby: CombyForm,
    credentials: CredentialsForm,
}

export const GenericCampaignForm: React.FunctionComponent<CampaignFormProps & { type: CampaignType }> = ({
    type,
    ...props
}) => {
    console.log(props.parsed)
    const Component = components[type]
    if (!props.campaignArguments) {
        props.onChange('{}')
        return null
    }
    if (!(props.parsed !== null && typeof props.parsed === 'object')) {
        return <ErrorAlert error="Invalid JSON: props was not an object. Please use the JSON editor to fix." />
    }
    return (
        <div className="mt-2">
            <Component {...props} />
        </div>
    )
}
