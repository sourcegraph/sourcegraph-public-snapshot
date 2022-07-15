import React, { useCallback } from 'react'

import { FlexTextArea, Input, Label } from '@sourcegraph/wildcard'

interface Props {
    value
    onChange: (newValue) => void
    disabled?: boolean
}

export const GitHubAppFormFields: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    value,
    onChange,
    disabled,
}) => {
    const onSlugChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            onChange({
                ...value,
                gitHubApp: {
                    slug: event.target.value,
                },
            })
        },
        [onChange, value]
    )
    const onAppIDChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => onChange({ ...value.gitHubApp, appID: event.target.value }),
        [onChange, value]
    )
    const onClientIDChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => onChange({ ...value.gitHubApp, clientID: event.target.value }),
        [onChange, value]
    )
    const onClientSecretChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => onChange({ ...value.gitHubApp, clientSecret: event.target.value }),
        [onChange, value]
    )
    const onPrivateKeyChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => onChange({ ...value.gitHubApp, privateKey: event.target.value }),
        [onChange, value]
    )
    return (
        <div data-testid="github-app-form-fields">
            <div className="form-group">
                <Label htmlFor="GitHubAppFormFields__slug">App Slug</Label>
                <Input
                    id="GitHubAppFormFields__slug"
                    className="test-GitHubAppFormFields-slug"
                    required={true}
                    aria-describedby="GitHubAppFormFields__slug-help"
                    onChange={onSlugChange}
                    value={value.gitHubApp.slug}
                />
                <small id="GitHubAppFormFields__slug-help" className="form-text text-muted">
                    The App Slug is the URL-friendly name of your GitHub App. You can find this on the settings page for
                    your GitHub App (e.g., https://github.com/settings/apps/:app_slug).
                </small>
            </div>
            <Input
                id="GitHubAppFormFields__appID"
                data-testid="test-GitHubAppFormFields__appID"
                disabled={disabled}
                spellCheck={false}
                placeholder="App ID"
                className="form-group"
                label="App ID"
                required={true}
                onChange={onAppIDChange}
                value={value.gitHubApp.appID}
            />

            <Input
                id="GitHubAppFormFields__clientID"
                disabled={disabled}
                spellCheck={false}
                placeholder="Client ID"
                className="form-group w-100"
                label="Client ID"
                required={true}
                onChange={onClientIDChange}
                value={value.gitHubApp.clientID}
            />

            <Input
                id="GitHubAppFormFields__clientSecret"
                disabled={disabled}
                spellCheck={false}
                placeholder="Client Secret"
                className="form-group w-100"
                label="Client Secret"
                required={true}
                onChange={onClientSecretChange}
                value={value.gitHubApp.clientSecret}
            />

            <div className="form-group">
                <Label htmlFor="GitHubAppFormFields__privateKey">Private Key</Label>
                <FlexTextArea
                    id="GitHubAppFormFields__privateKey"
                    minRows={27}
                    disabled={disabled}
                    spellCheck={false}
                    placeholder="Private Key"
                    className="form-group w-100"
                    required={true}
                    onChange={onPrivateKeyChange}
                    value={value.gitHubApp.privateKey}
                />
            </div>
        </div>
    )
}
