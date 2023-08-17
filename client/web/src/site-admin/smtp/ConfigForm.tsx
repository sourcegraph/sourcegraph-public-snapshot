import { FC, useCallback, useMemo, useState } from 'react'

import { ApolloError } from '@apollo/client'
import { mdiLink } from '@mdi/js'
import { applyEdits, modify, parse, ParseError } from 'jsonc-parser'

import { SiteConfiguration, SMTPServerConfig } from '@sourcegraph/shared/src/schema/site.schema'
import {
    AnchorLink,
    Button,
    Checkbox,
    Form,
    H3,
    Input,
    Label,
    Link,
    Alert,
    Select,
    Text,
    Container,
    Icon,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { LoaderButton } from '../../components/LoaderButton'
import { defaultModificationOptions } from '../SiteAdminConfigurationPage'

import { SendTestEmailForm } from './SendTestEmailForm'

interface Props {
    className?: string
    config?: string
    authenticatedUser: AuthenticatedUser
    saveConfig: (newContents: string) => Promise<void>
    loading?: boolean
    error?: ApolloError
}

interface FormData extends SMTPServerConfig {
    email?: SiteConfiguration['email.address']
    senderName?: SiteConfiguration['email.senderName']

    [key: string]: any
}

const initialConfig: FormData = {
    email: '',
    senderName: '',
    host: '',
    username: '',
    password: '',
    authentication: 'PLAIN',
    domain: '',
    port: 587,
    noVerifyTLS: false,
}

export const SMTPConfigForm: FC<Props> = ({ className, config, authenticatedUser, saveConfig, error, loading }) => {
    const [form, setForm] = useState<FormData>({ ...initialConfig })

    const [parsedConfig, err] = useMemo((): [FormData | null, Error | null] => {
        if (!config) {
            return [null, null]
        }
        const errors: ParseError[] = []
        const siteConfig = parse(config, errors, {
            allowTrailingComma: true,
            disallowComments: false,
        }) as SiteConfiguration

        if (errors?.length > 0) {
            const error = new Error('Cannot parse site config: ' + errors.join(', '))
            return [null, error]
        }

        const result = {
            email: siteConfig['email.address'] ?? '',
            senderName: siteConfig['email.senderName'] ?? '',
            ...siteConfig['email.smtp'],
            noVerifyTLS: !!siteConfig['email.smtp']?.noVerifyTLS,
            authentication: siteConfig['email.smtp']?.authentication ?? 'PLAIN',
        } as FormData

        setForm({
            ...initialConfig,
            ...result,
        })

        return [result, null]
    }, [config])

    const isValid = useMemo(
        () =>
            form.email &&
            form.host &&
            form.port &&
            (form.authentication === 'none' || (form.username && form.password)),
        [form]
    )

    const fieldRequired = useCallback(
        (field: string) => {
            if (!form[field]) {
                return `${field} is required`
            }
            return ''
        },
        [form]
    )

    const fieldChanged = useCallback(
        (evt: React.ChangeEvent<HTMLInputElement> | React.ChangeEvent<HTMLSelectElement>) => {
            const { name, value } = evt.target

            const newValue = {
                ...form,
                [name]: value,
            }
            if (name === 'noVerifyTLS') {
                newValue.noVerifyTLS = !(evt.target as HTMLInputElement).checked
            }
            if (name === 'port') {
                newValue.port = Number(value)
            }
            setForm(newValue)
        },
        [form, setForm]
    )

    const applyChanges = useCallback((): Promise<void> => {
        const normalizedConfig = { ...form } as FormData
        if (normalizedConfig.authentication === 'none') {
            delete normalizedConfig.username
            delete normalizedConfig.password
        }
        for (const [key, val] of Object.entries(normalizedConfig)) {
            if (val === '' || val === undefined) {
                delete normalizedConfig[key]
            }
        }

        let newConfig = applyEdits(
            config!,
            modify(config!, ['email.address'], normalizedConfig.email, defaultModificationOptions)
        )
        newConfig = applyEdits(
            newConfig,
            modify(newConfig!, ['email.senderName'], normalizedConfig.senderName, defaultModificationOptions)
        )
        newConfig = applyEdits(
            newConfig,
            modify(
                newConfig!,
                ['email.smtp'],
                {
                    host: normalizedConfig.host,
                    port: normalizedConfig.port,
                    authentication: normalizedConfig.authentication,
                    username: normalizedConfig.username,
                    password: normalizedConfig.password,
                    noVerifyTLS: normalizedConfig.noVerifyTLS,
                    domain: normalizedConfig.domain,
                },
                defaultModificationOptions
            )
        )

        return saveConfig(newConfig)
    }, [form, config, saveConfig])

    const reset = useCallback(() => {
        setForm({
            ...initialConfig,
            ...parsedConfig,
        })
    }, [parsedConfig])

    const handleSubmit = useCallback(
        (evt: React.FormEvent<HTMLFormElement>): Promise<void> => {
            evt.preventDefault()
            return applyChanges()
        },
        [applyChanges]
    )

    const effectiveError = err || error

    return (
        <div className={className}>
            <H3 id="smtp">
                SMTP Configuration{' '}
                <AnchorLink to="#smtp">
                    <Icon aria-label="link icon" svgPath={mdiLink} />
                </AnchorLink>
            </H3>
            <Text className="mt-2">
                Sourcegraph uses an SMTP server of your choosing to send emails.{' '}
                <Link to="/help/admin/config/email" target="_blank">
                    See documentation
                </Link>{' '}
                for more information.
            </Text>

            {effectiveError && <Alert variant="danger">{effectiveError.message}</Alert>}
            <Form onSubmit={handleSubmit} className="mt-2">
                <Label className="w-100 mt-2">
                    Email
                    <Input
                        name="email"
                        type="email"
                        message="The 'from' address for emails sent by this server."
                        value={form.email}
                        onChange={fieldChanged}
                        placeholder="noreply@sourcegraph.example.com"
                        error={fieldRequired('email')}
                    />
                </Label>
                <Label className="w-100 mt-2">
                    Sender name
                    <Input
                        name="senderName"
                        message="The name to use in the 'from' address for emails sent by this server."
                        value={form.senderName}
                        onChange={fieldChanged}
                    />
                </Label>
                <Label className="w-100 mt-2">
                    Host
                    <Input
                        name="host"
                        message="The hostname of the SMTP server that sends the email."
                        value={form.host}
                        onChange={fieldChanged}
                        placeholder="smtp.sourcegraph.example.com"
                        error={fieldRequired('host')}
                    />
                </Label>
                <Label className="w-100 mt-2">
                    Port
                    <Input
                        name="port"
                        type="number"
                        message="The port of the SMTP server that sends the email."
                        value={form.port}
                        onChange={fieldChanged}
                        placeholder="587"
                        error={fieldRequired('port')}
                    />
                </Label>
                <Label className="w-100 mt-2" id="auth-select-label">
                    Authentication
                </Label>
                <Select
                    aria-labelledby="auth-select-label"
                    name="authentication"
                    message="Authentication mechanism used to talk to SMTP server."
                    value={form.authentication}
                    onChange={fieldChanged}
                    className="mt-2"
                >
                    <option value="none">None</option>
                    <option value="PLAIN">Plain</option>
                    <option value="CRAM-MD5">Cram-MD5</option>
                </Select>
                {form.authentication !== 'none' && (
                    <>
                        <Label className="w-100 mt-2">
                            Username
                            <Input
                                name="username"
                                message="Username to authenticate with SMTP server."
                                value={form.username}
                                onChange={fieldChanged}
                                error={fieldRequired('username')}
                            />
                        </Label>
                        <Label className="w-100 mt-2">
                            {form.authentication === 'PLAIN' ? 'Password' : 'Secret'}
                            <Input
                                name="password"
                                type="password"
                                message="Password to authenticate with SMTP server."
                                value={form.password}
                                onChange={fieldChanged}
                                error={fieldRequired('password')}
                            />
                        </Label>
                    </>
                )}
                <Label className="w-100 mt-2">
                    Domain
                    <Input
                        name="domain"
                        message="The HELO domain to provide to the SMTP server (if needed)."
                        value={form.domain}
                        onChange={fieldChanged}
                    />
                </Label>
                <div className="mt-2">
                    <Checkbox
                        name="noVerifyTLS"
                        type="checkbox"
                        message="Enable/Disable TLS verification (if needed, by default ON)."
                        checked={!form.noVerifyTLS}
                        onChange={fieldChanged}
                        label="TLS verification"
                        id="no-verify-tls-checkbox"
                    />
                </div>
                <div className="mt-3 d-flex">
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        label="Save"
                        loading={loading}
                        disabled={!isValid || JSON.stringify(form) === JSON.stringify(parsedConfig)}
                    />
                    <Button className="ml-2" type="button" variant="secondary" onClick={reset}>
                        Discard changes
                    </Button>
                </div>
            </Form>
            <Container className="mt-4">
                <H3>Test email</H3>
                <SendTestEmailForm authenticatedUser={authenticatedUser} />
            </Container>
        </div>
    )
}
