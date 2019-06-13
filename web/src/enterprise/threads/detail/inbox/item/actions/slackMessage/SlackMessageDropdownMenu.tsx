import React, { useMemo, useState } from 'react'
import { DropdownItem, DropdownMenu, DropdownMenuProps } from 'reactstrap'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../../../shared/src/util/errors'

interface Props extends Pick<DropdownMenuProps, 'right'> {}

const LOADING: 'loading' = 'loading'

interface Recipient {
    label: string
    detail: string
}

const querySlackRecipients = async (): Promise<Recipient[]> => [
    { label: '@ijt', detail: 'author' },
    { label: '@farhan', detail: 'author' },
    { label: '@beyang', detail: 'code owner' },
    { label: '@christina', detail: 'code owner' },
    { label: '#search', detail: 'team' },
]

/**
 * A dropdown menu with a list of Slack message recipients.
 */
export const SlackMessageDropdownMenu: React.FunctionComponent<Props> = ({ ...props }) => {
    const [recipientsOrError, setRecipientsOrError] = useState<typeof LOADING | Recipient[] | ErrorLike>(LOADING)

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setRecipientsOrError(await querySlackRecipients())
        } catch (err) {
            setRecipientsOrError(asError(err))
        }
    }, [])

    const MAX_ITEMS = 9 // TODO!(sqs): hack

    return (
        <DropdownMenu {...props}>
            {recipientsOrError === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading Slack recipients...
                </DropdownItem>
            ) : isErrorLike(recipientsOrError) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading Slack recipients
                </DropdownItem>
            ) : (
                <>
                    <DropdownItem header={true} className="py-1">
                        Open in Slack...
                    </DropdownItem>
                    {recipientsOrError.slice(0, MAX_ITEMS).map(({ label, detail }, i) => (
                        <DropdownItem key={i} className="d-flex justify-content-between">
                            {label} <span className="text-muted ml-3">{detail}</span>
                        </DropdownItem>
                    ))}
                </>
            )}
        </DropdownMenu>
    )
}
