import { FC, useCallback, useMemo, useState } from 'react'

import { ApolloError } from '@apollo/client'
import Form from '@rjsf/core'
import { RJSFSchema } from '@rjsf/utils'
import validator from '@rjsf/validator-ajv8'
import { applyEdits, modify, parse, ParseError } from 'jsonc-parser'

import { SiteConfiguration, SMTPServerConfig } from '@sourcegraph/shared/src/schema/site.schema'
import { Button, Checkbox, H3, Icon, Input, Label, Link, Alert, Select, Text, Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { LoaderButton } from '../../components/LoaderButton'
import { theme } from '../schema-form/theme'
import { defaultModificationOptions } from '../SiteAdminConfigurationPage'

import { SendTestEmailForm } from './SendTestEmailForm'

const schema: RJSFSchema = {
    type: 'object',
    properties: {
        'email.smtp': {
            title: 'SMTPServerConfig',
            description:
                'The SMTP server used to send transactional emails.\nPlease see https://docs.sourcegraph.com/admin/config/email',
            type: 'object',
            additionalProperties: false,
            required: ['host', 'port', 'authentication'],
            properties: {
                host: {
                    description: 'The SMTP server host.',
                    type: 'string',
                },
                port: {
                    description: 'The SMTP server port.',
                    type: 'integer',
                },
                username: {
                    description: 'The username to use when communicating with the SMTP server.',
                    type: 'string',
                },
                password: {
                    description: 'The password to use when communicating with the SMTP server.',
                    type: 'string',
                },
                authentication: {
                    description: 'The type of authentication to use for the SMTP server.',
                    type: 'string',
                    enum: ['none', 'PLAIN', 'CRAM-MD5'],
                },
                domain: {
                    description: 'The HELO domain to provide to the SMTP server (if needed).',
                    type: 'string',
                },
                noVerifyTLS: {
                    description: 'Disable TLS verification',
                    type: 'boolean',
                },
                additionalHeaders: {
                    description:
                        "Additional headers to include on SMTP messages that cannot be configured with other 'email.smtp' fields.",
                    type: 'array',
                    items: {
                        title: 'Header',
                        type: 'object',
                        required: ['key', 'value'],
                        additionalProperties: false,
                        properties: {
                            key: {
                                type: 'string',
                            },
                            value: {
                                type: 'string',
                            },
                            sensitive: {
                                type: 'boolean',
                            },
                        },
                        examples: [
                            {
                                key: '',
                                value: '',
                            },
                        ],
                    },
                },
            },
            default: null,
            examples: [
                {
                    host: 'smtp.example.com',
                    port: 465,
                    username: 'alice',
                    password: 'mypassword',
                    authentication: 'PLAIN',
                },
            ],
            group: 'Email',
        },
        'email.address': {
            title: 'Email Address',
            description:
                'The "from" address for emails sent by this server.\nPlease see https://docs.sourcegraph.com/admin/config/email',
            type: 'string',
            format: 'email',
            group: 'Email',
            examples: ['noreply@sourcegraph.example.com'],
        },
        'email.senderName': {
            description: 'The name to use in the "from" address for emails sent by this server.',
            type: 'string',
            group: 'Email',
            default: 'Sourcegraph',
            examples: ['Our Company Sourcegraph', 'Example Inc Sourcegraph'],
        },
    },
}

interface Props {
    className?: string
    config?: string
    authenticatedUser: AuthenticatedUser
    saveConfig: (newContents: string) => Promise<void>
    loading?: boolean
    error?: ApolloError
}

interface FormData {
    ['email.address']?: SiteConfiguration['email.address']
    ['email.senderName']?: SiteConfiguration['email.senderName']
    ['email.smtp']?: SMTPServerConfig

    [key: string]: any
}

// const initialConfig: FormData = {
//     email: '',
//     senderName: '',
//     host: '',
//     username: '',
//     password: '',
//     authentication: 'PLAIN',
//     domain: '',
//     port: 587,
//     noVerifyTLS: false,
// }

export const SMTPConfigForm: FC<Props> = ({ className, config, authenticatedUser, saveConfig, error, loading }) => {
    const [form, setForm] = useState<FormData>({} as FormData)

    const [parsedConfig, err] = useMemo((): [FormData | null, Error | null] => {
        if (!config) {
            return [null, null]
        }
        let errors: ParseError[] = []
        const siteConfig = parse(config, errors, {
            allowTrailingComma: true,
            disallowComments: false,
        }) as SiteConfiguration

        if (errors?.length > 0) {
            const error = new Error('Cannot parse site config: ' + errors.join(', '))
            return [null, error]
        }

        const result = {
            ['email.address']: siteConfig['email.address'],
            ['email.senderName']: siteConfig['email.senderName'],
            ['email.smtp']: {
                ...siteConfig['email.smtp'],
            },
        }

        setForm({
            ...result,
        })

        return [result, null]
    }, [config])

    const isValid = useMemo(() => {
        return (
            form.email && form.host && form.port && (form.authentication === 'none' || (form.username && form.password))
        )
    }, [form])

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
        (e: React.ChangeEvent<HTMLInputElement> | React.ChangeEvent<HTMLSelectElement>) => {
            // const { name, value } = e.target
            // const newValue = {
            //     ...form,
            //     [name]: value,
            // }
            // if (name === 'noVerifyTLS') {
            //     newValue.noVerifyTLS = !(e.target as HTMLInputElement).checked
            // }
            // setForm(newValue)
        },
        [form, setForm]
    )

    const applyChanges = useCallback(() => {
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

        saveConfig(newConfig)
    }, [form, config, parsedConfig])

    const reset = useCallback(() => {
        setForm({
            ...initialConfig,
            ...parsedConfig,
        })
    }, [parsedConfig])

    const handleSubmit = useCallback(
        (e: React.FormEvent<HTMLFormElement>) => {
            e.preventDefault()
            applyChanges()
        },
        [applyChanges]
    )

    const formChanged = (data: IChangeEvent, id?: string) => {
        console.log('FORM CHANGED', data, id)
    }

    const effectiveError = err ?? error
    // const effectiveError = error

    return (
        <div className={className}>
            <H3>SMTP Configuration</H3>
            <Text className="mt-2">
                Sourcegraph uses an SMTP server of your choosing to send emails.{' '}
                <Link to="/help/admin/config/email" target="_blank">
                    See documentation
                </Link>{' '}
                for more information.
            </Text>

            {effectiveError && <Alert variant="danger">{effectiveError.message}</Alert>}
            <Form schema={schema} validator={validator} formData={form} theme={theme} onChange={formChanged}>
                <div className="mt-3 d-flex">
                    <LoaderButton type="submit" variant="primary" label="Save" loading={loading} />
                    {/* <Button className="ml-2" type="button" variant="secondary" onClick={reset}>
                        Discard changes
                    </Button> */}
                </div>
            </Form>
            <Container className="mt-4">
                <H3>Test email</H3>
                <SendTestEmailForm authenticatedUser={authenticatedUser} />
            </Container>
        </div>
    )
}
