import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'
import { FlexTextArea, Input, Label } from '@sourcegraph/wildcard'

export type GitHubAppFormFieldsValue = Pick<GQL.IUser, 'username' | 'displayName' | 'avatarURL'>

interface Props {
    disabled?: boolean
}

export const GitHubAppFormFields: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ disabled }) => (
    <div data-testid="github-app-form-fields">
        <div className="form-group">
            <Label htmlFor="GitHubAppFormFields__slug">App Slug</Label>
            <Input
                id="GitHubAppFormFields__slug"
                className="test-GitHubAppFormFields-slug"
                required={true}
                aria-describedby="GitHubAppFormFields__slug-help"
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
        />

        <Input
            id="GitHubAppFormFields__clientID"
            disabled={disabled}
            spellCheck={false}
            placeholder="Client ID"
            className="form-group w-100"
            label="Client ID"
            required={true}
        />

        <Input
            id="GitHubAppFormFields__clientSecret"
            disabled={disabled}
            spellCheck={false}
            placeholder="Client Secret"
            className="form-group w-100"
            label="Client Secret"
            required={true}
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
            />
        </div>
    </div>
)
