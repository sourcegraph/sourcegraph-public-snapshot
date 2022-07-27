import { mdiDownload } from '@mdi/js'
import lodash from 'lodash'

import { useLazyQuery } from '@sourcegraph/http-client'
import { Button, Icon } from '@sourcegraph/wildcard'

interface IProps<IQuery, IVariables> {
    query: string
    variables: IVariables
    path: string
    fileName: string
}

export function ExportButton<IQuery, IVariables>({
    query,
    variables,
    path,
    fileName,
}: IProps<IQuery, IVariables>): JSX.Element {
    const [fetchCSV] = useLazyQuery<IQuery, IVariables>(query, {
        variables,
        onCompleted: data => {
            const element = document.createElement('a')
            element.setAttribute('href', 'data:text/csv;charset=utf-8,' + encodeURIComponent(lodash.get(data, path)))
            element.setAttribute('download', `${fileName}.csv`)
            element.style.display = 'none'
            document.body.append(element)
            element.click()
            element.remove()
        },
    })

    return (
        <Button variant="primary" onClick={() => fetchCSV()}>
            <Icon className="mr-1" color="var(--white)" svgPath={mdiDownload} size="sm" aria-label="Download icon" />{' '}
            Export
        </Button>
    )
}
