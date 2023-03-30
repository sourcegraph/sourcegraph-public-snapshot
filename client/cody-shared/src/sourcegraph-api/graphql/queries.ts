export const CURRENT_USER_ID_QUERY = `
query CurrentUser {
    currentUser {
        id
    }
}`

export const REPOSITORY_ID_QUERY = `
query Repository($name: String!) {
	repository(name: $name) {
		id
	}
}`

export const SEARCH_EMBEDDINGS_QUERY = `
query EmbeddingsSearch($repo: ID!, $query: String!, $codeResultsCount: Int!, $textResultsCount: Int!) {
	embeddingsSearch(repo: $repo, query: $query, codeResultsCount: $codeResultsCount, textResultsCount: $textResultsCount) {
		codeResults {
			fileName
			startLine
			endLine
			content
		}
		textResults {
			fileName
			startLine
			endLine
			content
		}
	}
}`

export const IS_CONTEXT_REQUIRED_QUERY = `
query IsContextRequiredForChatQuery($query: String!) {
	isContextRequiredForChatQuery(query: $query)
}`

export const LOG_EVENT_MUTATION = `
mutation {
    logEvent(
    event: String!
    userCookieID: String!
    url: String!
    source: EventSource!
    argument: String
    publicArgument: String
    ) {
    alwaysNil
    }
}`
