interface ParsedLabelProps {
    filter: string
    label: string
}
export const ParsedLabel: React.FunctionComponent<React.PropsWithChildren<ParsedLabelProps>> = ({ filter, label }) => {
    if (filter.length === 0) {
        return <span>{label}</span>
    }

    const matcher = new RegExp(`(${filter})`, 'ig')
    const splitLabel = label.split(matcher)

    return (
        <>
            {splitLabel.map((chunk, index) =>
                // Splitting a string will have no acceptable keys
                // eslint-disable-next-line react/no-array-index-key
                matcher.test(chunk) ? <b key={index}>{chunk}</b> : <span key={index}>{chunk}</span>
            )}
        </>
    )
}
