import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { useParams } from 'react-router-dom'

import { ErrorMessage, LoadingSpinner } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'

import { useSpongeLog } from './backend'

export const SpongeLog: React.FunctionComponent<{}> = ({}) => {
    const { uuid } = useParams<{ uuid: string }>()
    if (uuid === undefined) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle="UUID undefined LOLOLO" />
    }
    const { data, loading, error } = useSpongeLog(uuid)
    if (loading) {
        return <LoadingSpinner />
    }
    if (error) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={error} />} />
    }
    // This is custom for autocomplete - should later dispatch on data.spongeLog.interpreter.
    const log = JSON.parse(data?.spongeLog?.log || 'null')
    // Return JSON pretty printed in a <code> tag
    return (
        <pre>
            <code>{JSON.stringify(log, null, 2)}</code>
        </pre>
    )
}
