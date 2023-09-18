import { type FC, useState, type Dispatch, type SetStateAction } from 'react'

import classnames from 'classnames'

import { Button, H2, Input, Text, Modal } from '@sourcegraph/wildcard'

import styles from './GetStarted.module.scss'

export function GetStarted(): JSX.Element {
    const [isOpen, setIsOpen] = useState(true)
    const [repoInput, setRepoInput] = useState('')
    const [emailInput, setEmailInput] = useState('')
    const [languageInput, setLanguageInput] = useState([])

    const [step, setStep] = useState(0)
    const repoModal: JSX.Element = (
        <ModalContent
            input={repoInput}
            onHandleInput={setRepoInput}
            name="repository"
            placeholder="Enter the repository name"
            title="Before we start, what repository do you work in most?"
            label="Example: react-website or host domain/organization/repository-name"
            step={1}
            onHandleNext={setStep}
        />
    )
    const emailModal: JSX.Element = (
        <ModalContent
            input={emailInput}
            onHandleInput={setEmailInput}
            name="email"
            placeholder="Enter an email address"
            title="What email are your commits associated with?"
            label="Example: person@company.com"
            step={2}
            onHandleNext={setStep}
        />
    )

    const languageModal: JSX.Element = (
        <ModalContentButtons
            title="What language do you write the most?"
            onHandleInput={setLanguageInput}
            step={3}
            onHandleNext={setStep}
        />
    )

    return (
        <Modal
            className={styles.modal}
            containerClassName={styles.modalContainer}
            onDismiss={() => {}}
            position="center"
            aria-labelledby="license-key"
        >
            {step === 0 && repoModal}
            {step === 1 && emailModal}
            {step === 2 && languageModal}
        </Modal>
    )
}

const availableLanguages = ['TypeScript', 'C', 'Go', 'Java', 'Php', 'Python', 'JavaScript']

interface ModalContentButtonsProps {
    title: string
    onHandleInput: Dispatch<SetStateAction<never[]>>
    step: number
    onHandleNext: Dispatch<SetStateAction<number>>
}

const ModalContentButtons: FC<ModalContentButtonsProps> = ({
    title,
    onHandleInput,
    step,
    onHandleNext,
}): JSX.Element => (
    <div className={styles.fade}>
        <H2 className={styles.title}>{title}</H2>
        <div className={classnames(styles.wrapper, styles.languageWrapper)}>
            {availableLanguages.map((language, index) => (
                <span
                    key={index}
                    className={styles.language}
                    onClick={() => onHandleInput(prev => [...prev, language])}
                >
                    {language}
                </span>
            ))}
        </div>
        <div className={styles.container}>
            <div className={styles.step}>{`${step} of 3`}</div>
            <Button className={styles.button} onClick={() => onHandleNext(prevState => prevState + 1)}>
                Next
            </Button>
        </div>
    </div>
)

interface ModalContentProps {
    input: string
    placeholder: string
    label: string
    onHandleInput: Dispatch<SetStateAction<string>>
    title: string
    name: string
    step: number
    onHandleNext: Dispatch<SetStateAction<number>>
}

const ModalContent: FC<ModalContentProps> = ({
    input,
    title,
    onHandleInput,
    placeholder,
    label,
    name,
    step,
    onHandleNext,
}): JSX.Element => (
    <div className={styles.fade}>
        <H2 className={styles.title}>{title}</H2>
        <div className={styles.wrapper}>
            <Input
                type="text"
                name={name}
                value={input}
                className="pb-3"
                onChange={({ target: { value } }) => onHandleInput(value)}
                placeholder={placeholder}
            />
            <Text className={styles.text}>{label}</Text>
        </div>
        <div className={styles.container}>
            <div className={styles.step}>{`${step} of 3`}</div>
            <Button className={styles.button} onClick={() => onHandleNext(prevState => prevState + 1)}>
                Next
            </Button>
        </div>
    </div>
)
