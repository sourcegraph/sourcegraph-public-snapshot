import { basicSyntaxColumns } from '$lib/branded'

export function getQueryExamples(): { title: string; columns: typeof basicSyntaxColumns }[] {
    return [
        {
            title: 'Code search basics',
            columns: basicSyntaxColumns,
        },
    ]
}
