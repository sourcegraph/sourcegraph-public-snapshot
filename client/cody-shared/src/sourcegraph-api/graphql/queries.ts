export const CURRENT_USER_ID_QUERY = `
query CurrentUser {
    currentUser {
        id
    }
}`

export const CURRENT_SITE_VERSION_QUERY = `
query SiteProductVersion {
    site {
        productVersion
    }
}`

export const CURRENT_SITE_HAS_CODY_ENABLED_QUERY = `
query SiteHasCodyEnabled {
    site {
        isCodyEnabled
    }
}`

export const CURRENT_SITE_GRAPHQL_FIELDS_QUERY = `
query SiteGraphQLFields {
    __type(name: "Site") {
        fields {
            name
        }
    }
}`

export const CURRENT_USER_ID_AND_VERIFIED_EMAIL_QUERY = `
query CurrentUser {
    currentUser {
        id
        hasVerifiedEmail
    }
}`

export const REPOSITORY_ID_QUERY = `
query Repository($name: String!) {
	repository(name: $name) {
		id
	}
}`

export const REPOSITORY_EMBEDDING_EXISTS_QUERY = `
query Repository($name: String!) {
	repository(name: $name) {
                id
                embeddingExists
	}
}`

export const SEARCH_EMBEDDINGS_QUERY = `
query EmbeddingsSearch($repos: [ID!]!, $query: String!, $codeResultsCount: Int!, $textResultsCount: Int!) {
	embeddingsMultiSearch(repos: $repos, query: $query, codeResultsCount: $codeResultsCount, textResultsCount: $textResultsCount) {
		codeResults {
            repoName
            revision
			fileName
			startLine
			endLine
			content
		}
		textResults {
            repoName
            revision
			fileName
			startLine
			endLine
			content
		}
	}
}`

export const LEGACY_SEARCH_EMBEDDINGS_QUERY = `
query LegacyEmbeddingsSearch($repo: ID!, $query: String!, $codeResultsCount: Int!, $textResultsCount: Int!) {
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

export const SEARCH_TYPE_REPO_QUERY = `
query SearchTypeRepo($query: String!) {
	search(query: $query, version: V3) {
        results {
            limitHit
            results {
                ... on Repository {
                    name
                }
            }
        }
	}
}`

export const IS_CONTEXT_REQUIRED_QUERY = `
query IsContextRequiredForChatQuery($query: String!) {
	isContextRequiredForChatQuery(query: $query)
}`

export const LOG_EVENT_MUTATION = `
mutation LogEventMutation($event: String!, $userCookieID: String!, $url: String!, $source: EventSource!, $argument: String, $publicArgument: String) {
    logEvent(
		event: $event
		userCookieID: $userCookieID
		url: $url
		source: $source
		argument: $argument
		publicArgument: $publicArgument
    ) {
		alwaysNil
	}
}`
