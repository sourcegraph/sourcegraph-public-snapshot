import { FC, useCallback, useEffect, useMemo, useState } from 'react'

import { SiteConfiguration, SMTPServerConfig } from '@sourcegraph/shared/src/schema/site.schema'
import { Checkbox, Form, Input, Label, Link, Select, Text, useDebounce } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'

import { SendTestEmailForm } from './SendTestEmailForm'

type EmailConfiguration = Pick<SiteConfiguration, 'email.address' | 'email.senderName' | 'email.smtp'>
interface Props {
    config?: EmailConfiguration
    authenticatedUser: AuthenticatedUser
    onConfigChange: (newConfig: EmailConfiguration) => void
}

interface FormData extends SMTPServerConfig {
    email?: EmailConfiguration['email.address']
    senderName?: EmailConfiguration['email.senderName']

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

export const SMTPConfigForm: FC<Props> = ({ config, authenticatedUser, onConfigChange }) => {
    const [form, setForm] = useState<FormData>({ ...initialConfig })

    useEffect(() => {
        if (!config) {
            return
        }

        const result = {
            email: config['email.address'] ?? '',
            senderName: config['email.senderName'] ?? '',
            ...config['email.smtp'],
            noVerifyTLS: !!config['email.smtp']?.noVerifyTLS,
            authentication: config['email.smtp']?.authentication ?? 'PLAIN',
        } as FormData

        setForm({
            ...initialConfig,
            ...result,
        })
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

    const applyChanges = useCallback(() => {
        if (!isValid) {
            return
        }

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

        onConfigChange({
            'email.address': normalizedConfig.email,
            'email.senderName': normalizedConfig.senderName,
            'email.smtp': {
                host: normalizedConfig.host,
                port: normalizedConfig.port,
                authentication: normalizedConfig.authentication,
                username: normalizedConfig.username,
                password: normalizedConfig.password,
                noVerifyTLS: normalizedConfig.noVerifyTLS,
                domain: normalizedConfig.domain,
            },
        })
    }, [form, onConfigChange, isValid])

    const applyChangesDebounced = useDebounce(applyChanges, 300)

    return (
        <>
            <Text className="mt-2">
                Sourcegraph uses an SMTP server of your choosing to send emails.{' '}
                <Link to="/help/admin/config/email" target="_blank">
                    See documentation
                </Link>{' '}
                for more information.
            </Text>

            <Form className="mt-2" onChange={applyChangesDebounced}>
                <Label className="w-100 mt-2">
                    <Text className="mb-2">Email</Text>
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
                    <Text className="mb-2">Sender name</Text>
                    <Input
                        name="senderName"
                        message="The name to use in the 'from' address for emails sent by this server."
                        value={form.senderName}
                        onChange={fieldChanged}
                    />
                </Label>
                <Label className="w-100 mt-2">
                    <Text className="mb-2">Host</Text>
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
                    <Text className="mb-2">Port</Text>
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
                >
                    <option value="none">None</option>
                    <option value="PLAIN">Plain</option>
                    <option value="CRAM-MD5">Cram-MD5</option>
                </Select>
                {form.authentication !== 'none' && (
                    <>
                        <Label className="w-100 mt-2">
                            <Text className="mb-2">Username</Text>
                            <Input
                                name="username"
                                message="Username to authenticate with SMTP server."
                                value={form.username}
                                onChange={fieldChanged}
                                error={fieldRequired('username')}
                            />
                        </Label>
                        <Label className="w-100 mt-2">
                            <Text className="mb-2">{form.authentication === 'PLAIN' ? 'Password' : 'Secret'}</Text>
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
                    <Text className="mb-2">Domain</Text>
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
            </Form>
            <SendTestEmailForm authenticatedUser={authenticatedUser} className="mt-4" />
        </>
    )
}
