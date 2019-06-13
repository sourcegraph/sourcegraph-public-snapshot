/*
// TODO!(sqs): use relative path/rev for DiscussionThreadTargetRepo
const queryInboxItems = (threadID: GQL.ID): Promise<GQL.IDiscussionThreadTargetConnection> =>
    queryGraphQL(
        gql`
            query ThreadInboxItems($threadID: ID!) {
                node(id: $threadID) {
                    __typename
                    ... on DiscussionThread {
                        targets {
                            nodes {
                                __typename
                                ...DiscussionThreadTargetFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
            }
            ${discussionThreadTargetFieldsFragment}
        `,
        { threadID }
    )
        .pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.node ||
                    data.node.__typename !== 'DiscussionThread' ||
                    !data.node.targets ||
                    !data.node.targets.nodes
                ) {
                    throw createAggregateError(errors)
                }
                return data.node.targets
            })
        )
        .toPromise()
*/
/*

    const [, setItems0OrError] = useState<
        | typeof LOADING
        | (GQL.IDiscussionThreadTargetConnection & { matchingNodes: GQL.IDiscussionThreadTargetRepo[] })
        | ErrorLike
    >(LOADING)

		
				    // tslint:disable-next-line: no-floating-promises
    useEffectAsync(async () => {
        try {
            const data = await queryInboxItems(thread.id)
            const isHandled = (item: GQL.IDiscussionThreadTargetRepo): boolean =>
                (threadSettings.pullRequests || []).some(pull => pull.items.includes(item.id))
            setItems0OrError({
                ...data,
                matchingNodes: data.nodes
                    .filter(
                        (item): item is GQL.IDiscussionThreadTargetRepo =>
                            item.__typename === 'DiscussionThreadTargetRepo'
                    )
                    .filter(
                        item =>
                            (query.includes('is:open') && !item.isIgnored && !isHandled(item)) ||
                            (query.includes('is:ignored') && item.isIgnored && !isHandled(item)) ||
                            (!query.includes('is:open') && !query.includes('is:ignored'))
                    )
                    .filter(item => {
                        const m = query.match(/repo:([^\s]+)/)
                        if (m && m[1]) {
                            const repo = m[1]
                            const ids = (threadSettings.pullRequests || [])
                                .filter(pull => pull.repo === repo)
                                .flatMap(pull => pull.items)
                            return ids.includes(item.id)
                        }
                        return true
                    }),
            })
        } catch (err) {
            setItems0OrError(asError(err))
        }
    }, [thread.id, threadSettings])

		*/
