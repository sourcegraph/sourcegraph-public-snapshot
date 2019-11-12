import * as jsonc from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import * as _monaco from 'monaco-editor'

import { truncate } from 'lodash'
import * as React from 'react'
import { from, Subscription, timer } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { switchMap } from 'rxjs/operators'
import './CriticalConfigEditor.scss'
import { MonacoEditor } from './MonacoEditor'

/**
 * Amount of time to wait before showing the loading indicator.
 */
const WAIT_BEFORE_SHOWING_LOADER = 250 // ms

// TODO(slimsag): future: Warn user if they are discarding changes
// TODO(slimsag): future: Explicit discard changes button?
// TODO(slimsag): future: Better button styling
// TODO(slimsag): future: Better link styling
// TODO(slimsag): future: Better 'loading' state styling

/**
 * The success response from the API /get and /update endpoints.
 */
interface Configuration {
    /**
     * The unique ID of this configuration version.
     */
    ID: string

    /**
     * The literal JSONC configuration.
     */
    Contents: string
}

/**
 * The parameters that mut be POST to the /update endpoint.
 */
interface UpdateParams {
    /**
     * The last Configuration.ID value the client was aware of. If outdated,
     * the update will fail.
     */
    LastID: string

    /**
     * The literal JSONC configuration.
     */
    Contents: string
}

interface Props {}

interface State {
    /** The current config content according to the server. */
    criticalConfig: Configuration | null

    /** The current content in the editor. */
    content: string | null

    /** Whether or not the loader can be shown yet, iff criticalConfig === null */
    canShowLoader: boolean

    /** Whether or not the Monaco editor has loaded */
    hasLoadedEditor: boolean

    /** Whether or not to show a "Saving..." indicator */
    showSaving: boolean

    /** Whether or not to show a "Saved!" indicator */
    showSaved: boolean

    /** Whether or not to show a saving error indicator */
    showSavingError: string | null
}

/** A response from the server when an error occurs. */
export interface ErrorResponse {
    /** A human-readable error message. */
    error: string
    /** A stable ID for this kind of error. */
    code: string
}

const defaultFormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const quickConfigureActions: {
    id: string
    label: string
    run: (config: string) => { edits: jsonc.Edit[]; selectText: string }
}[] = [
    {
        id: 'setExternalURL',
        label: 'Set external URL',
        run: config => {
            const value = '<external URL>'
            const edits = setProperty(config, ['externalURL'], value, defaultFormattingOptions)
            return { edits, selectText: '<external URL>' }
        },
    },
    {
        id: 'setLicenseKey',
        label: 'Set license key',
        run: config => {
            const value = '<license key>'
            const edits = setProperty(config, ['licenseKey'], value, defaultFormattingOptions)
            return { edits, selectText: '<license key>' }
        },
    },
    {
        id: 'addGitLabAuth',
        label: 'Add GitLab sign-in',
        run: config => {
            const edits = setProperty(
                config,
                ['auth.providers', -1],
                {
                    type: 'gitlab',
                    displayName: 'GitLab',
                    url: '<GitLab URL>',
                    clientID: '<client ID>',
                    clientSecret: '<client secret>',
                },
                defaultFormattingOptions
            )
            return { edits, selectText: '<GitLab URL>' }
        },
    },
    {
        id: 'addGitHubAuth',
        label: 'Add GitHub sign-in',
        run: config => {
            const edits = setProperty(
                config,
                ['auth.providers', -1],
                {
                    type: 'github',
                    displayName: 'GitHub',
                    url: 'https://github.com/',
                    allowSignup: true,
                    clientID: '<client ID>',
                    clientSecret: '<client secret>',
                },
                defaultFormattingOptions
            )
            return { edits, selectText: '<client ID>' }
        },
    },
    {
        id: 'useOneLoginSAML',
        label: 'Add OneLogin SAML',
        run: config => {
            const { externalURL, externalURLRegexp } = getExternalURLPlaceholders(config)
            const value = {
                type: 'saml',
                displayName: 'OneLogin',
                COMMENT_1: true,
                COMMENT_2: true,
                identityProviderMetadataURL: '<identity provider metadata URL>',
            }
            const comments = {
                COMMENT_1: `
      // OneLogin SAML instructions
      // ==========================
      //
      // Before proceeding, ensure you've set externalURL to the appropriate value.
      // (The instructions below use the current value of externalURL.)
      //
      // Create a SAML app in OneLogin:
      // 1. Go to https://mycompany.onelogin.com/apps/find (replace "mycompany" with your
      //    company's OneLogin ID).
      // 2. Select "SAML Test Connector (SP)" and click "Save".
      // 3. Under the "Configuration" tab, set the following properties:
      //    Audience:  ${externalURL}/.auth/saml/metadata
      //    Recipient: ${externalURL}/.auth/saml/acs
      //    ACS (Consumer) URL Validator: ${externalURLRegexp}\\/\\.auth\\/saml\\/acs
      //    ACS (Consumer) URL: ${externalURL}/.auth/saml/acs
      // 4. Under the "Parameters" tab, ensure the following parameters exist:
      //    Email (NameID): Email
      //    DisplayName:    First Name         Include in SAML Assertion: ✓
      //    login:          AD user name       Include in SAML Assertion: ✓
      // 5. Save the app in OneLogin and fill in the fields below:`,
                COMMENT_2: `
      // This URL describes OneLogin to Sourcegraph. Find it in the OneLogin app config GUI
      // under the "SSO" tab, under "Issuer URL".
      // It should look something like "https://mycompany.onelogin.com/saml/metadata/123456"
      // or "https://app.onelogin.com/saml/metadata/123456".`,
            }
            const edits = [editWithComments(config, ['auth.providers', -1], value, comments)]
            return { edits, selectText: 'OneLogin SAML instructions' }
        },
    },
    {
        id: 'useOktaSAML',
        label: 'Add Okta SAML',
        run: config => {
            const { externalURL } = getExternalURLPlaceholders(config)
            const value = {
                type: 'saml',
                displayName: 'Okta',
                COMMENT_1: true,
                COMMENT_2: true,
                identityProviderMetadataURL: '<identity provider metadata URL>',
            }
            const comments = {
                COMMENT_1: `
      // Okta SAML instructions
      // ======================
      //
      // Before proceeding, ensure you've set externalURL to the appropriate value.
      // (The instructions below use the current value of externalURL.)
      //
      // Create a SAML app in Okta:
      // 1. Go to the Okta admin "Add Application" page, classic UI (looks like
      //    https://my-org.okta.com/admin/apps/add-app or https://dev-12345.oktapreview.com/admin/apps/add-app).
      // 2. Click "Create New App", select "SAML 2.0", and "Create".
      // 3. Give the app the name "Sourcegraph", click "Next".
      // 4. Set the following SAML settings:
      //    Single Sign On URL: ${externalURL}/.auth/saml/acs
      //      Use this for Recipient URL and Destination URL: ✓
      //    Audience URI (SP Entity ID) / Audience Restriction: ${externalURL}/.auth/saml/metadata
      //    Attribute statements:
      //      Email: user.email
      //      Login: user.login
      //      DisplayName: \${user.firstName} \${user.lastName}
      //    Click "Next".
      // 5. Select "I'm an Okta customer adding an internal app" and click "Finish".
      // 6. Go to the "Assignments" tab, click the "Assign" dropdown > "Assign to Groups"
      //    > Everyone ("Assign" button).
      // 7. Fill in the fields below:`,
                COMMENT_2: `
      // This URL describes Okta to Sourcegraph. Go to the "Sign On" tab and copy
      // the hyperlink "Identity Provider metadata is available if this application
      // supports dynamic configuration." The value looks like
      // "https://my-org.okta.com/app/abcdefghijk012345678/sso/saml/metadata" or "https://dev-123435.oktapreview.com/app/abcdefghijk012345678/sso/saml/metadata".`,
            }
            const edits = [editWithComments(config, ['auth.providers', -1], value, comments)]
            return { edits, selectText: 'Okta SAML instructions' }
        },
    },
    {
        id: 'useSAML',
        label: 'Add other SAML',
        run: config => {
            const edits = setProperty(
                config,
                ['auth.providers', -1],
                {
                    type: 'saml',
                    displayName: 'SAML',
                    identityProviderMetadataURL: '<SAML IdP metadata URL>',
                },
                defaultFormattingOptions
            )
            return { edits, selectText: '<SAML IdP metadata URL>' }
        },
    },
    {
        id: 'useOIDC',
        label: 'Add OpenID Connect',
        run: config => {
            const edits = setProperty(
                config,
                ['auth.providers', -1],
                {
                    type: 'openidconnect',
                    displayName: 'OpenID Connect',
                    issuer: '<identity provider URL>',
                    clientID: '<client ID>',
                    clientSecret: '<client secret>',
                },
                defaultFormattingOptions
            )
            return { edits, selectText: '<identity provider URL>' }
        },
    },
]

export class CriticalConfigEditor extends React.PureComponent<Props, State> {
    public state: State = {
        criticalConfig: null,
        content: null,
        canShowLoader: false,
        showSaving: false,
        showSaved: false,
        showSavingError: null,
        hasLoadedEditor: false,
    }

    private configEditor?: _monaco.editor.ICodeEditor

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Periodically rerender our component in case our request takes longer
        // than `WAIT_BEFORE_SHOWING_LOADER` and we need to show the loading
        // indicator.
        this.subscriptions.add(timer(WAIT_BEFORE_SHOWING_LOADER).subscribe(t => this.setState({ canShowLoader: true })))

        // Load the initial critical config.
        this.subscriptions.add(
            fromFetch('/api/get')
                .pipe(
                    switchMap(response => {
                        if (response.status !== 200) {
                            throw new Error(`Error saving: ${response.status} ${response.statusText}`)
                        }
                        return response.json()
                    })
                )
                .subscribe({
                    next: (config: Configuration) => {
                        this.setState({
                            criticalConfig: config,
                            content: config.Contents,
                        })
                    },
                    error: error => {
                        console.error(error)
                        alert(error.message) // TODO(slimsag): Better general error state here.
                        return
                    },
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private runAction(id: string, editor?: _monaco.editor.ICodeEditor): void {
        if (editor) {
            const action = editor.getAction(id)
            action.run().then(
                () => undefined,
                (err: any) => console.error(err)
            )
        }
    }

    public static isStandaloneCodeEditor(
        editor: _monaco.editor.ICodeEditor
    ): editor is _monaco.editor.IStandaloneCodeEditor {
        return editor.getEditorType() === _monaco.editor.EditorType.ICodeEditor
    }

    public render(): JSX.Element | null {
        const actions = quickConfigureActions
        return (
            <div className="critical-config-editor">
                {actions && this.state.hasLoadedEditor && (
                    <div className="critical-config-editor__action-groups">
                        <div className="critical-config-editor__action-group-header">Quick configure:</div>
                        <div className="critical-config-editor__actions">
                            {actions.map(({ id, label }) => (
                                <button
                                    key={id}
                                    className="btn btn-secondary btn-sm critical-config-editor__action"
                                    onClick={() => this.runAction(id, this.configEditor)}
                                    type="button"
                                >
                                    {label}
                                </button>
                            ))}
                        </div>
                    </div>
                )}

                <div
                    className={`critical-config-editor__monaco-reserved-space${
                        this.state.criticalConfig ? ' critical-config-editor__monaco-reserved-space--monaco' : ''
                    }`}
                >
                    {!this.state.criticalConfig && this.state.canShowLoader && <div>Loading...</div>}

                    {this.state.criticalConfig && (
                        <MonacoEditor
                            content={this.state.criticalConfig.Contents}
                            language="json"
                            onDidContentChange={this.onDidContentChange}
                            onDidSave={this.onDidSave}
                            editorWillMount={this.editorWillMount}
                        />
                    )}
                </div>
                <button type="button" onClick={this.onDidSave}>
                    Save changes
                </button>
                {this.state.showSaving && <span className="critical-config-editor__status-indicator">Saving...</span>}
                {this.state.showSaved && (
                    <span className="critical-config-editor__status-indicator critical-config-editor__status-indicator--success">
                        Saved!
                    </span>
                )}
                {this.state.showSavingError && (
                    <span className="critical-config-editor__status-indicator critical-config-editor__status-indicator--error">
                        {this.state.showSavingError}
                    </span>
                )}
            </div>
        )
    }

    /**
     * Private helper that stores a reference to the Monaco editor after it's mounted.
     * This is used to run the "Quick configure" actions.
     */
    private editorWillMount = (editor: _monaco.editor.IStandaloneCodeEditor, model: _monaco.editor.IModel): void => {
        this.configEditor = editor
        if (CriticalConfigEditor.isStandaloneCodeEditor(editor)) {
            for (const { id, label, run } of quickConfigureActions) {
                editor.addAction({
                    label,
                    id,
                    run: editor => {
                        editor.focus()
                        editor.pushUndoStop()
                        const { edits, selectText } = run(editor.getValue())
                        const monacoEdits = toMonacoEdits(model, edits)
                        let selection: _monaco.Selection | undefined
                        if (typeof selectText === 'string') {
                            const afterText = jsonc.applyEdits(editor.getValue(), edits)
                            let offset = afterText.slice(edits[0].offset).indexOf(selectText)
                            if (offset !== -1) {
                                offset += edits[0].offset
                                selection = _monaco.Selection.fromPositions(
                                    getPositionAt(afterText, offset),
                                    getPositionAt(afterText, offset + selectText.length)
                                )
                            }
                        }
                        if (!selection) {
                            // TODO: This is buggy. See
                            // https://github.com/sourcegraph/sourcegraph/issues/2756.
                            selection = _monaco.Selection.fromPositions(
                                monacoEdits[0].range.getStartPosition(),
                                monacoEdits[monacoEdits.length - 1].range.getEndPosition()
                            )
                        }
                        editor.executeEdits(id, monacoEdits, [selection])
                        editor.revealPositionInCenter(selection.getStartPosition())
                    },
                })
            }
        }
        this.setState({ hasLoadedEditor: true })
    }

    private onDidContentChange = (content: string): void => this.setState({ content })

    private onDidSave = (): void => {
        this.setState(
            {
                showSaving: true,
                showSaved: false,
                showSavingError: null,
            },
            () =>
                this.subscriptions.add(
                    from(
                        fetch('/api/update', {
                            method: 'POST',
                            body: JSON.stringify({
                                LastID: this.state.criticalConfig!.ID,
                                Contents: this.state.content,
                            } as UpdateParams),
                        })
                            .then(async response => {
                                if (response.status !== 200) {
                                    const text = await response.text()
                                    const truncatedText = truncate(text, { length: 30 })
                                    return {
                                        error: `Unexpected HTTP ${response.status}: ${truncatedText}`,
                                    }
                                }
                                return response.json()
                            })
                            .catch(error => ({
                                error:
                                    error instanceof TypeError && error.message === 'Failed to fetch'
                                        ? 'Network error - check the browser console for details'
                                        : `error: ${error}`,
                            }))
                    ).subscribe((response: { error: any } | Configuration) => {
                        if ('error' in response) {
                            this.setState({
                                showSaving: false,
                                showSaved: false,
                                showSavingError: response.error.toString(),
                            })
                            return
                        }
                        this.setState({
                            criticalConfig: response,
                            content: response.Contents,
                            showSaving: false,
                            showSaved: true,
                            showSavingError: null,
                        })

                        // Hide the saved indicator after 2.5s.
                        setTimeout(() => this.setState({ showSaved: false }), 2500)
                    })
                )
        )
    }
}

function toMonacoEdits(
    model: _monaco.editor.IModel,
    edits: jsonc.Edit[]
): _monaco.editor.IIdentifiedSingleEditOperation[] {
    return edits.map((edit, i) => ({
        identifier: { major: model.getVersionId(), minor: i },
        range: _monaco.Range.fromPositions(
            model.getPositionAt(edit.offset),
            model.getPositionAt(edit.offset + edit.length)
        ),
        forceMoveMarkers: true,
        text: edit.content,
    }))
}

/**
 * Returns the (line, column) position into the text (@param text) at the given
 * byte offset (@param offset).
 */
function getPositionAt(text: string, offset: number): _monaco.IPosition {
    const lines = text.split('\n')
    let pos = 0
    let i = 0
    for (const line of lines) {
        if (offset < pos + line.length + 1) {
            return new _monaco.Position(i + 1, offset - pos + 1)
        }
        pos += line.length + 1
        i++
    }
    throw new Error(`offset ${offset} out of bounds in text of length ${text.length}`)
}

/**
 * editWithComment returns a Monaco edit action that sets the value of a JSON field with a
 * "//" comment annotating the field. The comment is inserted wherever
 * `"COMMENT_SENTINEL": true` appears in the JSON.
 */
function editWithComments(
    config: string,
    path: jsonc.JSONPath,
    value: any,
    comments: { [key: string]: string }
): jsonc.Edit {
    const edit = setProperty(config, path, value, defaultFormattingOptions)[0]
    for (const commentKey of Object.keys(comments)) {
        edit.content = edit.content.replace(`"${commentKey}": true,`, comments[commentKey])
        edit.content = edit.content.replace(`"${commentKey}": true`, comments[commentKey])
    }
    return edit
}

function getExternalURLPlaceholders(config: string): { externalURL: string; externalURLRegexp: string } {
    let externalURL
    let externalURLRegexp
    try {
        externalURL = jsonc.parse(config).externalURL
        externalURLRegexp = externalURL.replace(/\//g, '\\/')
    } catch {
        /* not necessarily an error, config might be empty */
    }
    if (!externalURL) {
        externalURL = '<externalURL>'
        externalURLRegexp = '<externalURL regex>'
    }
    return { externalURL, externalURLRegexp }
}
