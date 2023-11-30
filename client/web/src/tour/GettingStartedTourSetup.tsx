import {
    type FC,
    useState,
    type PropsWithChildren,
    useRef,
    type ChangeEvent,
    useLayoutEffect,
    useEffect,
    type FormEvent,
} from 'react'

import { useApolloClient } from '@apollo/client'
import { useDebounce } from 'use-debounce'

import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    Button,
    H2,
    Input,
    Text,
    Modal,
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
    type TetherInstanceAPI,
    Flipping,
    Form,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { LoaderButton } from '../components/LoaderButton'
import type { UserOnboardingRepoValidationResult, UserOnboardingRepoValidationVariables } from '../graphql-operations'
import { useLanguageCompletionSource, useRepositoryCompletionSource } from '../search/autocompletion/hooks'
import { eventLogger } from '../tracking/eventLogger'

import styles from './GettingStartedTourSetup.module.scss'

const DIALOG_TITLE_ID = 'onboarding-setup-title'

interface GettingStartedTourSetupProps {
    user: AuthenticatedUser
}

export const GettingStartedTourSetup: FC<GettingStartedTourSetupProps> = ({ user }) => {
    const [open, setOpen] = useState(true)
    const [repoInput, setRepoInput] = useState('')
    const [emailInput, setEmailInput] = useState('')
    const [languageInput, setLanguageInput] = useState('')
    const [step, setStep] = useState(0)
    const [, setConfig] = useTemporarySetting('onboarding.userconfig')

    const nextStep = (): void => setStep(step => step + 1)
    const done = (): void => {
        eventLogger.log('TourSetupCompleted')
        setOpen(false)
    }

    useEffect(() => {
        if (open) {
            eventLogger.log('TourSetupShown')
        }
    }, [open])

    useEffect(() => {
        if (!open && repoInput && emailInput && languageInput) {
            setConfig({
                skipped: false,
                userinfo: {
                    repo: repoInput,
                    email: emailInput,
                    language: languageInput,
                },
            })
        }
    }, [open, repoInput, emailInput, languageInput, setConfig])

    const steps: ((step: number) => React.ReactElement)[] = [
        step => <RepositoryModal step={[step + 1, steps.length]} onSelect={setRepoInput} onHandleNext={nextStep} />,
        step => (
            <EmailModal step={[step + 1, steps.length]} onSelect={setEmailInput} onHandleNext={nextStep} user={user} />
        ),
        step => (
            <LanguageModal
                step={[step + 1, steps.length]}
                onSelect={setLanguageInput}
                onHandleNext={done}
                repo={repoInput}
            />
        ),
    ]

    if (!open) {
        return null
    }

    return (
        <Modal
            className={styles.modal}
            containerClassName={styles.modalContainer}
            onDismiss={() => {}}
            aria-labelledby={DIALOG_TITLE_ID}
        >
            {steps[step](step)}
        </Modal>
    )
}

interface ModalInnerProps {
    title: string
    step: [number, number]
    label?: string
    loading?: boolean
    onHandleNext?: () => void
}

const ModalInner: FC<PropsWithChildren<ModalInnerProps>> = ({
    label,
    title,
    step: [step, totalSteps],
    onHandleNext,
    loading,
    children,
}): JSX.Element => {
    const [, setConfig] = useTemporarySetting('onboarding.userconfig')
    const onSubmit = (event: FormEvent): void => {
        event.preventDefault()
        onHandleNext?.()
    }
    const skip = (): void => {
        eventLogger.log('TourSetupSkipped')
        setConfig({ skipped: true })
    }
    return (
        <Form onSubmit={onSubmit}>
            <div className={styles.fade}>
                <H2 id={DIALOG_TITLE_ID} className={styles.title}>
                    {title}
                </H2>
                <div className={styles.wrapper}>
                    {children}
                    {label && <Text className="mt-3 text-muted">{label}</Text>}
                </div>
                <div className={styles.container}>
                    <div className="text-muted">{`${step} of ${totalSteps}`}</div>
                    <Button variant="link" onClick={skip}>
                        Skip
                    </Button>
                    <LoaderButton
                        type="submit"
                        variant="merged"
                        className={styles.button}
                        disabled={!onHandleNext}
                        alwaysShowLabel={true}
                        loading={loading}
                        label={step === totalSteps ? 'Done' : 'Next'}
                    />
                </div>
            </div>
        </Form>
    )
}

interface ModalContentProps extends Pick<ModalInnerProps, 'step' | 'onHandleNext'> {
    onSelect: (value: string) => void
}

const REPO_QUERY = gql`
    query UserOnboardingRepoValidation($name: String) {
        repository(name: $name) {
            name
        }
    }
`

const RepositoryModal: FC<ModalContentProps> = ({ step, onHandleNext, onSelect }) => {
    const [value, setValue] = useState('')
    const [isValidating, setIsValidating] = useState(false)
    const [error, setError] = useState('')
    const [debouncedSearchTerm] = useDebounce(value, 100)
    const { suggestions } = useRepositoryCompletionSource(debouncedSearchTerm)

    const client = useApolloClient()

    const handleSearchTermChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setValue(event.target.value)
    }

    async function validateRepo(): Promise<void> {
        const repo = value
        if (!repo.trim()) {
            setError('Please enter a repository name.')
            return
        }

        setIsValidating(true)
        const result = await client.query<UserOnboardingRepoValidationResult, UserOnboardingRepoValidationVariables>({
            query: getDocumentNode(REPO_QUERY),
            variables: { name: value },
        })
        // Only proceed if the value hasn't changed in the meantime
        if (value === repo) {
            if (result.error) {
                setError(result.error.message)
            } else if (!result.data.repository) {
                setError(`Repository '${repo}' doesn't exist. Please enter a valid repository name.`)
            } else {
                onSelect(repo)
                onHandleNext?.()
            }
            setIsValidating(false)
        }
    }

    // This is an (unfortunate) workaround for making the combobox's arrow keys
    // work properly when data is loaded in asynchronously.
    // By using a different key we are forcing the list component to reinitialize.
    // Without it pressing arrow keys will jump to seemingly random elements in the
    // list.
    const [listKey, setListKey] = useState(0)
    useLayoutEffect(() => {
        setListKey(key => key + 1)
    }, [suggestions])

    // This is needed to ensure that the popover is corrently positioned and sized
    // as async data is coming in.
    const tetherRef = useRef<TetherInstanceAPI>()
    useLayoutEffect(() => {
        tetherRef.current?.forceUpdate()
    }, [tetherRef, suggestions])

    return (
        <ModalInner
            title="Before we start, what repository do you work in most?"
            label="Example: react-website or  host domain/organization/repository-name"
            step={step}
            onHandleNext={value.trim() && !isValidating && !error ? validateRepo : undefined}
            loading={isValidating}
        >
            <Combobox aria-label="Choose a repo" openOnFocus={true} hidden={false} onSelect={setValue}>
                <ComboboxInput
                    autoFocus={true}
                    spellCheck={false}
                    autocomplete={true}
                    autoComplete="off"
                    placeholder="Enter repository name"
                    onInput={handleSearchTermChange}
                    onFocus={() => setError('')}
                    error={error}
                    name="repository"
                    required={true}
                />

                <ComboboxPopover flipping={Flipping.opposite} onTetherCreate={tether => (tetherRef.current = tether)}>
                    <ComboboxList key={listKey}>
                        {suggestions.map(suggestion => (
                            <ComboboxOption key={suggestion} value={suggestion} />
                        ))}
                    </ComboboxList>
                </ComboboxPopover>
            </Combobox>
        </ModalInner>
    )
}

interface EmailModalProps extends ModalContentProps {
    user: AuthenticatedUser
}

const EmailModal: FC<EmailModalProps> = ({ step, onHandleNext, onSelect, user }) => {
    const [email, setEmail] = useState(user.emails.find(email => email.isPrimary)?.email ?? '')
    const [error, setError] = useState('')

    const input = useRef<HTMLInputElement>(null)

    function validate(): void {
        const message = input.current?.validationMessage
        if (message) {
            setError(message)
        } else {
            onSelect(email)
            onHandleNext?.()
        }
    }

    return (
        <ModalInner
            title="What email are your commits associated with?"
            label="Example: person@company.com"
            step={step}
            onHandleNext={email && !error ? validate : undefined}
        >
            <Input
                ref={input}
                name="email"
                type="email"
                title="Enter your commit email address"
                placeholder="Enter an email address"
                autoFocus={true}
                required={true}
                value={email}
                onInput={(event: ChangeEvent<HTMLInputElement>) => setEmail(event.target.value)}
                onFocus={() => setError('')}
                error={error}
            />
        </ModalInner>
    )
}

interface LanguageModalProps extends ModalContentProps {
    repo: string
}

const LanguageModal: FC<LanguageModalProps> = ({ step, onHandleNext, repo, onSelect }) => {
    const [language, setLanguage] = useState('')
    const [error, setError] = useState('')
    const { suggestions } = useLanguageCompletionSource(language)

    const input = useRef<HTMLInputElement>(null)

    function validate(): void {
        const message = input.current?.validationMessage
        if (message) {
            setError(message)
        } else {
            const languageLower = language.toLowerCase()
            if (!suggestions.some(suggestion => suggestion.toLowerCase() === languageLower)) {
                setError(`Unknown language '${language}'. Please select one from the suggested languages.`)
            } else {
                onSelect(language)
                onHandleNext?.()
            }
        }
    }

    return (
        <ModalInner
            title={`What language do you use the most in ${displayRepoName(repo)}?`}
            step={step}
            onHandleNext={language && !error ? validate : undefined}
        >
            <Combobox
                className="mt-3"
                aria-label="Choose a repo"
                openOnFocus={true}
                hidden={false}
                onSelect={setLanguage}
            >
                <ComboboxInput
                    autoFocus={true}
                    spellCheck={false}
                    autocomplete={true}
                    autoComplete="off"
                    placeholder="Enter language name"
                    onInput={(event: ChangeEvent<HTMLInputElement>) => setLanguage(event.target.value)}
                    onFocus={() => setError('')}
                    error={error}
                    name="repository"
                    required={true}
                />

                <ComboboxPopover flipping={Flipping.opposite}>
                    <ComboboxList>
                        {suggestions.map(suggestion => (
                            <ComboboxOption key={suggestion} value={suggestion} />
                        ))}
                    </ComboboxList>
                </ComboboxPopover>
            </Combobox>
        </ModalInner>
    )
}
