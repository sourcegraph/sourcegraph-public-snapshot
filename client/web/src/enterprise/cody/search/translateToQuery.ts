import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { CompletionRequest, getCodyCompletionOneShot } from '../api'

export async function translateToQuery(input: string, user: AuthenticatedUser | null): Promise<string | null> {
    const messages = getCompletionRequestMessages(input, user)
    const result = await getCodyCompletionOneShot(messages)
    if (!result.includes('contents>') && !result.includes('filters>')) {
        return null
    }
    const query = result
        .replace('<contents>', ' ')
        .replace('</contents>', ' ')
        .replace('<filters>', ' ')
        .replace('</filters>', ' ')
        .replace(/\n/g, ' ')
        .replace(/\s{2,}/g, ' ')
        .trim()
    return query
}

const getCompletionRequestMessages = (input: string, user: AuthenticatedUser | null): CompletionRequest['messages'] => [
    {
        speaker: 'human',
        text:
            `I am a sourcegraph user and my username is ${user?.username || 'anonymous'}.\n` +
            'You are an expert at writing queries that match what a human requests. A typical query describes content or filenames to find across all repositories using the following structure: <contents>REGEXP</contents><filters>FILTER FILTER FILTER...</fitlers>.\n' +
            'REGEXP is a regular expression used to match file content in the repositories. REGEXP in put inside the <contents> block. You may leave the <contents></contents> block empty.\n' +
            'FILTER has the following structure tag:value. You can put multiple FILTER inside the <filters> block like <filters>repo:my_repo lang:typescript type:symbol</fitlers>. FILTER is used to limit search to specific repositories or files. Multiple or combined FILTER are intersected.' +
            'The following FILTERS are available for writing the queries:\n' +
            '1. <filters>repo:REPO-NAME-REGEXP</filters>. Only include results from repositories whose path matches the REPO-NAME-REGEXP.\n' +
            '2. <filters>repo:REPO-NAME-REGEXP@REVISION</filters> or <filters>repo:REPO-NAME-REGEXP rev:REVISION</filters>. The REVISION part refers to repository revisions (branches, commit hashes, and tags) and may take on the following forms: @branch (a branch name), @1735d48 (a commit hash), @3.15 (a tag). If you want results from specific branch my_branch on a repo use <fitlers>repo:REPO-NAME-REGEXP@my_branch</fitlers> or <filter>repo:REPO-NAME-REGEXP rev:my_branch</filters>. If you want results from specific commit 1735d48 in a repo use <filters>repo:REPO-NAME-REGEXP@1735d48</fitlers> or <filters>repo:REPO-NAME-REGEXP rev:1735d48</fitlers>.\n' +
            '3. <filters>-repo:REPO-NAME-REGEXP</filters>. It excludes results from repositories whose path matches the REPO-NAME-REGEXP. It can be used in combination of repo:REPO-NAME-REGEXP filter. If you want to include frontend repo and exclude backend repo use <filters>repo:frontend -repo:backend</filters>.\n' +
            '4. <filters>file:PATH-REGEXP</filters>. Only include results in files whose full path matches the PATH_REGEXP. Always use the <filters>file:PATH-REGEXP</filters> filter to narrow the query to only specific files. Escape any special characters in the regular expression to match file contents (such as \\* to match a literal *)\n' +
            '5. <filters>-file:PATH-REGEXP</filters>. It excludes results in files whose full path matches the PATH_REGEXP. It can be used in combination of file:PATH-REGEXP filter. If you want to include \\.go files and exclude _test\\.go files use <filters>file:\\.go -file:_test\\.go</filters>.\n' +
            '6. <filters>file:has.owner(OWNER_NAME)</filters>. Only include results from files owned by a user with name matching OWNER_NAME. TO only include files owner by james-connor use <filters>file.has.owner(james-connor).\n' +
            '7. <filters>file:has.owner(OWNER_EMAIL)</filters>. Only include results from files owned by a user with email matching OWNER_EMAIL. TO only include files owner by peter@gmail.com use <filters>file.has.owner(peter@gmail.com).\n' +
            '8. <filters>lang:LANGUAGE</filters>. Only include results from files in the specified LANGUAGE. LANGUAGE can be typescript, javascript, go, css, scss, html, markdown, rust, c++, java, etc.\n' +
            '9. <filters>type:SEARCH_TYPE</filters>. SEARCH_TYPE Specifies the type of search. If you want to search over changes to code use type:diff. If you want to search over a commit message use type:commit. If you want to search for a symbol use type:symbol.\n' +
            '10. <filters>type:diff</filters>. seaches over changes performed including adding code, deleting code, deleting files, new file created, removal of code etc. If you want to search over any such changes use type:diff.\n' +
            '11. <filters>author:"AUTHOR_PATTERN"</filters>. Only include results from diffs or commits authored by AUTHOR_PATTERN. Regexps are supported for AUTHOR_PATTERN. AUTHOR_PATTERN is of the form "Full Name <user@example.com>", where Full Name is name of the author and user@example.com is the email of the author. To include commits from authors from a specific domain, use <filters>type:commit author:"example.com$"</filters>. To include commits from author having USERNAME as name use <filters>type:commit author:"USERNAME"</filters>. Do not escape special characters in AUTHOR_PATTERN.\n' +
            '12. <filters>author:@SOURCEGRAPH_USERNAME</filters>. Only include results from diffs or commits authored by the Sourcegraph user having either email or username matching with SOURCEGRAPH_USERNAME. SOURCEGRAPH_USERNAME is the username of user in the sourcegraph instance. To include diffs or commits from sourcegraph user having username as SG_USER use <filters>type:commit author:@SG_USER</filters>.\n' +
            '13. <filters>-author:"AUTHOR_PATTERN"</filters>. Excludes results from diffs or commits authored by AUTHOR_PATTERN. Regexps are supported for AUTHOR_PATTERN. AUTHOR_PATTERN is of the form "Full Name <user@example.com>", where Full Name is name of the author and user@example.com is the email of the author. To exclude commits from authors from a specific domain, use <filters>type:commit -author:"example.com>$"</filters>. To exclude diffs or commits from author having USERNAME as name but include commit from author having ANOTHER_USER as name use <filters>type:commit author:"ANOTHER_USER" -author:"USERNAME"</filters>.\n' +
            '14. <filters>-author:@SOURCEGRAPH_USERNAME</filters>. Exclude results from commits authored by the Sourcegraph user having either email or username matching with SOURCEGRAPH_USERNAME. SOURCEGRAPH_USERNAME is the username of user in the sourcegraph instance. To exclude diffs or commits from sourcegraph user having username as SG_USER but include commits from sourcegraph user having username as ANOTHER_SG_USER use <filters>type:commit author:@ANOTHER_SG_USER -author:@SG_USER</filters>.\n' +
            '15. <filters>before:"TIME_FRAME"</filters> Only include results from diffs or commits which have a commit date before the specified TIME_FRAME. To only include diff or commits before last thursday use <filters>before:"last thursday"</filters>. To only include diffs or commits before 1 november 2020 use <filters>type:before:"1 november 2020".\n' +
            '16. <filters>after:"TIME_FRAME"</filters> Only include results from diffs or commits which have a commit date after the specified TIME_FRAME. To only include diff or commits after last thursday use <filters>after:"last thursday"</filters>. To only include diffs or commits after 1 november 2020 use <filters>after:"1 november 2020".\n' +
            '17. <filters>message:"COMMIT_MESSAGE"</filters> Only include results from diffs or commits which have a commit message containing the string COMMIT_MESSAGE. To include diffs or commits including bugfix use <fitlers>message:"bugfix"</filters>.\n' +
            '18. <filters>select:repo</filters> displays only distinct repo paths from the search results. If you want to list repos having package.json file use <filters>select:repo file:package.json</filters>.\n' +
            '19. <filters>select:file.owners</filters> displays owners of files from the search results. If you want to find owners of ci directory use <filters>select:file.owners file:ci/</filters>.\n' +
            '20. <filters>select:file</filters> displays file paths from the search results. If you want to list all files inside backend directory use <filters>select:file file:backend/</filters>.\n' +
            '21. <filters>select:commit.diff.added</filters> when searching commit diffs, select only diffs where the pattern matches on added lines. For example to search for recent commits that added TODOs in your code use <contents>TODO</contents><filters>type:diff select:commit.diff.added</filters>. This filter can be used to search for added lines of code, added references to symbols, added function calls etc.\n' +
            '22. <filters>select:commit.diff.removed</filters> when searching commit diffs, select only diffs where the pattern matches on removed lines. For example to search for recent commits that removed TODOs in your code use <contents>TODO</contents><filters>type:diff select:commit.diff.removed</filters>. This filter can be used to search for removed lines of code, removed references to symbols, removed function calls etc.\n' +
            '23. <filters>select:symbol.SYMBOL_KIND</filters> only finds symbols of specific kind as SYMBOL_KIND based on the search contents. SYMBOL_KIND can be of following types: file, module, namespace, package, class, method, property, field, constructor, enum, interface, function, variable, constant, string, number, boolean, array, object, key, null, enum-number, struct, event, operator, type-parameter. For example to find functions with name like use.* they query will be <contents>use.*</contents><filters>select.symbol.function type:symbol</filters>. Always use type:symbol only when using select:symbol.SYMBOL_KIND. To list all symbols of a particular kind without any pattern matching use <contents>.*</contents><filters>select:symbol.SYMBOL_KIND type:symbol</filters>, for example to list all classes in typescript files use <contents>.*</contents><filters>select:symbol.class type:symbol lang:typescript</filters>.\n\n' +
            'There are operators as well which helps to contruct query that human requests. There are 2 operators AND, OR, NOT.\n' +
            'AND operator returns results for files containing matches on the left and right side of the AND. To search files containing both foo and bar use <contents>foo AND bar</contents>.\n' +
            'OR returns file content matching either on the left or right side, or both. To search files containing both foo or bar use <contents>foo OR bar</contents>.\n' +
            'NOT can be used in place of - to negate keywords, such as file, content, lang, repohasfile, and repo. For search patterns, NOT excludes documents that contain the term after NOT. For readability, you can also include the AND operator before a NOT (i.e. panic NOT ever is equivalent to panic AND NOT ever).\n' +
            'If the search pattern content has a string which literally conflicts with syntax of any of the fitlers mentioned above, like <contents>repo:analytics</contents> use <filters>content:"repo:analytics"</filters> instead.',
    },
    { speaker: 'assistant', text: 'Understood. I will follow these rules.' },

    // repo filter basic
    { speaker: 'human', text: 'What is the query for <request>multierror repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>repo:multierror</filters>' },

    // repo filter with branch name tag
    { speaker: 'human', text: 'What is the query for <request>branch user/bug-fix on react repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>repo:react@user/bug-fix</filters>' },

    // repo filter with commit tag
    { speaker: 'human', text: 'What is the query for <request>commit 1g246ih in express repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>repo:express@1g246ih</filters>' },

    // excluding repo fitler
    {
        speaker: 'human',
        text: 'What is the query for <request>featureFlags and flagSets excluding dotenv repo</request>?',
    },
    { speaker: 'assistant', text: '<contents>featureFlags AND flagSets</contents><filters>-repo:dotenv</filters>' },

    // both excluding & including repo fitler
    { speaker: 'human', text: 'What is the query for <request>all microsoft repo excluding windows98</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>repo:microsoft -repo:windows98</filters>' },

    // file filter with npm package file name
    { speaker: 'human', text: 'What is the query for <request>npm packages that depend on react</request>?' },
    { speaker: 'assistant', text: '<contents>"react"</contents><filters>file:package\\.json</filters>' },

    // file filter with file path regexp
    { speaker: 'human', text: 'What is the query for <request>styles/**.scss files</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:styles/**\\.scss</filters>' },

    // file filter with directory regexp
    { speaker: 'human', text: 'What is the query for <request>files in styles directory</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:styles</filters>' },

    // file filter with go test files specific directory regexp
    {
        speaker: 'human',
        text: 'what is the query for <request>go test files in the client directory that contain the string "openid" or "auth"</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>openid OR auth</contents><filters>file:client/ file:_test\\.go$</filters>',
    },

    // file filter with graphql query files specific directory regexp
    {
        speaker: 'human',
        text: 'what is the query for <request>graphql files in the backend directory that includes currentUser: User"</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>currentUser: User</contents><filters>file:backend/ file:\\.graphql$</filters>',
    },

    // file filter with owner name tag
    { speaker: 'human', text: 'What is the query for <request>owned by ryan@outlook.com</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:has.owner(ryan@outlook.com)</filters>' },

    // file filter with exclude owner
    { speaker: 'human', text: 'What is the query for <request>exclude files owned by ryan@outlook.com</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>-file:has.owner(ryan@outlook.com)</filters>' },

    // file filter with owner tag and diff
    { speaker: 'human', text: 'What is the query for <request>changes to go files owned by alice</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:diff lang:go file:has.owner(alice)</filters>' },

    // file filter with owner tag react storybook specific
    { speaker: 'human', text: 'What is the query for <request>React storybook files owned by alice</request>?' },
    { speaker: 'assistant', text: '<contents>@storybook/react</contents><filters>file:has.owner(alice)</filters>' },

    // file filter with repo
    { speaker: 'human', text: 'What is the query for <request>hooks/use.*\\.tsx files in nextjs repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:hooks/use.*\\.tsx repo:nextjs</filters>' },

    // file filter with dir and exluding repo
    { speaker: 'human', text: 'What is the query for <request>pages dir excluding vue repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:pages/ -repo:vue</filters>' },

    // file filter with directory inside repo
    { speaker: 'human', text: 'What is the query for <request>components directory in frontend repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>file:components repo:frontend/</filters>' },

    // excluding file fitler
    { speaker: 'human', text: 'What is the query for <request>NewUser excluding go test files</request>?' },
    { speaker: 'assistant', text: '<contents>NewUser</contents><filters>-file:_test\\.go$</filters>' },

    // both including and excluding file filter
    {
        speaker: 'human',
        text: 'What is the query for <request>all scss files in styles folder excluding global\\.scss file</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>file:styles/.*\\.scss$ -file:global\\.scss lang:scss</filters>',
    },

    // diff filter
    { speaker: 'human', text: 'What is the query for <request>changes to authentication</request>?' },
    { speaker: 'assistant', text: '<contents>authentication</contents><filters>type:diff</filters>' },

    // diff filter for repo with commit tag
    {
        speaker: 'human',
        text: 'What is the query for <request>diff under commit 6k3j5ip on tensorflow repository</request>?',
    },
    { speaker: 'assistant', text: '<contents></contents><filters>repo:tensorflow@6k3j5ip type:diff</filters>' },

    // diff filter with nested file
    {
        speaker: 'human',
        text: 'What is the query for <request>changes made to .btn in global\\.css in styles directory</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>.btn</contents><filters>type:diff lang:css file:styles/.*global\\.css</filters>',
    },

    // diff filter with language
    { speaker: 'human', text: 'What is the query for <request>diff for rust files</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:diff lang:rust</filters>' },

    // diff filter with file directory
    { speaker: 'human', text: 'What is the query for <request>changes inside config/aws directory</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:diff file:config/aws/</filters>' },

    // diff filter with file directory and query
    { speaker: 'human', text: 'What is the query for <request>changes to .button under styles directory</request>?' },
    { speaker: 'assistant', text: '<contents>.button</contents><filters>type:diff file:styles/</filters>' },

    // diff filter for package.json with before
    { speaker: 'human', text: 'What is the query for <request>node packages changes before last friday</request>?' },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>type:diff file:package\\.json before:"last friday"</filters>',
    },

    // diff filter for symbol with before
    {
        speaker: 'human',
        text: 'What is the query for <request>changes to useAuthenticatedUser before september 17 1998</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>useAuthenticatedUser</contents><filters>type:diff before:"september 17 1998"</filters>',
    },

    // diff filter for nested file with after
    {
        speaker: 'human',
        text: 'What is the query for <request>diffs for file init\\.ts inside server dir after 3 weeks ago</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>type:diff file:server/.*init\\.ts after:"3 weeks ago"</filters>',
    },

    // diff filter for symbol with after
    {
        speaker: 'human',
        text: 'What is the query for <request>changes to useAuthenticatedUser after september 17 1998</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>useAuthenticatedUser</contents><filters>type:diff after:"september 17 1998"</filters>',
    },

    // diff filter with owner
    { speaker: 'human', text: 'What is the query for <request>changes to files owned by jordan</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:diff file.has.owner(jordan)</filters>' },

    // author filter with pattern
    {
        speaker: 'human',
        text: 'What is the query for <request>modifications made to go routes file by mrnugget</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>file:routes\\.go type:diff author:"mrnugget"</filters>',
    },

    // author filter with file and diff
    {
        speaker: 'human',
        text: 'What is the query for <request>changes to schema.graphql file by Rahul Gandhi <rahul@gmail.com></request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>file:schema.graphql type:diff author:"Rahul Gandhi <rahul@gmail.com>"</filters>',
    },

    // author filter with search and diff
    { speaker: 'human', text: 'What is the query for <request>changes to useLocalStorage by @ryan</request>?' },
    { speaker: 'assistant', text: '<contents>useLocalStorage</contents><filters>type:diff author:@ryan</filters>' },

    // author filter with me in nested dir
    {
        speaker: 'human',
        text: 'What is the query for <request>changes made to go files in graphqlbackend dir by me</request>?',
    },
    {
        speaker: 'assistant',
        text: `contents></contents><filters>file:graphqlbackend/.*\\.go type:diff author:@${
            user?.username || 'anonymous'
        }</filters>`,
    },

    // including authors from domain but excluding myself
    {
        speaker: 'human',
        text: 'What is the query for <request>changes made to schema.graphql by users having email domain gmail\\.com excluding myself</request>?',
    },
    {
        speaker: 'assistant',
        text: `contents></contents><filters>file:schema.graphql type:diff author:"@gmail.com$" -author:@${
            user?.username || 'anonymous'
        }</filters>`,
    },

    // commits by authors from domain but excluding pattern
    {
        speaker: 'human',
        text: 'What is the query for <request>commits by users from domain outlook\\.com excluding @kyle</request>?',
    },
    {
        speaker: 'assistant',
        text: 'contents></contents><filters>type:commit author:"@outlook.com$" -author:@kyle</filters>',
    },

    // commit filter with message
    { speaker: 'human', text: 'What is the query for <request>bugfix commits</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:commit message:"bugfix"</filters>' },

    // latest commits correction response, otherwise it adds order:desc on its own
    { speaker: 'human', text: 'What is the query for <request>latest commits by @freddy</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>type:commit author:@freddy</filters>' },

    // symbol filter with query
    { speaker: 'human', text: 'What is the query for <request>isAuthenticated symbol</request>?' },
    { speaker: 'assistant', text: '<contents>isAuthenticated</contents><filters>type:symbol</filters>' },

    // symbol filter with filePath
    { speaker: 'human', text: 'What is the query for <request>getUserByID symbol in models/users\\.rb</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>getUserByID</contents><filters>file:models/users\\.rb type:symbol</filters>',
    },

    // symbol filter with repo
    { speaker: 'human', text: 'What is the query for <request>Auth symbol in rails repo</request>?' },
    { speaker: 'assistant', text: '<contents>Auth</contents><filters>repo:rails type:symbol</filters>' },

    // symbol filter with excluding repo
    { speaker: 'human', text: 'What is the query for <request>passport symbol exluding node repo</request>?' },
    { speaker: 'assistant', text: '<contents>passport</contents><filters>-repo:node type:symbol</filters>' },

    // symbol filter with both including and excluding repo
    {
        speaker: 'human',
        text: 'What is the query for <request>Middleware symbol in express repo exluding node repo</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>Middleware</contents><filters>repo:express -repo:node type:symbol</filters>',
    },

    // symbol filter vs diff filter
    { speaker: 'human', text: 'What is the query for <request>changes to login symbol in javascript file</request>?' },
    { speaker: 'assistant', text: '<contents>login</contents><filters>type:diff lang:javascript</filters>' },

    // language fitler
    { speaker: 'human', text: 'What is the query for <request>typescript files</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>lang:go</filters>' },

    // language fitler with query
    { speaker: 'human', text: 'What is the query for <request>golang oauth</request>?' },
    { speaker: 'assistant', text: '<contents>oauth</contents><filters>lang:go</filters>' },

    // language fitler with react hook specific query
    { speaker: 'human', text: 'What is the query for <request>TypeScript files that define a React hook</request>' },
    {
        speaker: 'assistant',
        text: '<contents>^export (const|function) use\\w+</contents><filters>lang:typescript</filters>',
    },

    // language fitler with react class component specific query
    { speaker: 'human', text: 'What is the query for <request>react class components</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>class \\w+ extends React\\.Component</contents><filters>(lang:typescript OR lang:javascript)</filters>',
    },

    // language fitler with repo
    { speaker: 'human', text: 'What is the query for <request>python files in numpy repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>lang:python repo:numpy</filters>' },

    // language filter with excluding repo
    { speaker: 'human', text: 'What is the query for <request>scss files excluding tailwind repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>lang:scss -repo:tailwind</filters>' },

    // select repo with files filter
    { speaker: 'human', text: 'What is the query for <request>repos with node_modules</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:repo file:node_modules/</filters>' },

    // select repo with files filter
    { speaker: 'human', text: 'What is the query for <request>repository with \\.github</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:repo file:\\.github/</filters>' },

    // select repo with text
    { speaker: 'human', text: 'What is the query for <request>repositories with powered by cody</request>?' },
    { speaker: 'assistant', text: '<contents>powered by cody</contents><filters>select:repo</filters>' },

    // select repo with symbol search
    { speaker: 'human', text: 'What is the query for <request>repositories using Sentry symbol</request>?' },
    { speaker: 'assistant', text: '<contents>Sentry</contents><filters>select:repo type:symbol</filters>' },

    // select repo with commit author
    { speaker: 'human', text: 'What is the query for <request>repos with commits by mark</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:repo type:commit author:"mark"</filters>' },

    // select repo with name
    { speaker: 'human', text: 'What is the query for <request>google/* repos</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:repo repo:google/*</filters>' },

    // select owners with files filter
    { speaker: 'human', text: 'What is the query for <request>owners of schema\\.graphql</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:file.owners file:schema\\.graphql/</filters>' },

    // select owners with text
    { speaker: 'human', text: 'What is the query for <request>who owns files with useLocalStorage</request>?' },
    { speaker: 'assistant', text: '<contents>useLocalStorage</contents><filters>select:file.owners</filters>' },

    // select owners with symbol search
    { speaker: 'human', text: 'What is the query for <request>file owners having DashboardLayout symbol</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>DashboardLayout</contents><filters>select:file.owners type:symbol</filters>',
    },

    // select owners with revision filter
    {
        speaker: 'human',
        text: 'What is the query for <request>owners of files changed under commit 4569f55</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>select:file.owners type:commit rev:4569f55"</filters>',
    },

    // select owners with files filter
    { speaker: 'human', text: 'What is the query for <request>people owning enterprise/utils</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:file.owners file:enterprise/utils</filters>' },

    // select owners with repo filter
    { speaker: 'human', text: 'What is the query for <request>owners of lodash repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:file.owners repo:lodash</filters>' },

    // select owners with file & repo filter
    { speaker: 'human', text: 'What is the query for <request>owners of assets dir in adminui repo</request>?' },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>select:file.owners file:assets/ repo:adminui</filters>',
    },

    // select files with repo filter
    { speaker: 'human', text: 'What is the query for <request>api/* files in microservice repo</request>?' },
    { speaker: 'assistant', text: '<contents></contents><filters>select:file file:api/* repo:microservice</filters>' },

    // select python files
    { speaker: 'human', text: 'What is the query for <request>python files</request>?' },
    { speaker: 'assistant', text: '<contents>powered by cody</contents><filters>select:file file:\\.py</filters>' },

    // select files with symbol search
    { speaker: 'human', text: 'What is the query for <request>files using useEffect(</request>?' },
    { speaker: 'assistant', text: '<contents>useEffect(</contents><filters>select:file type:symbol</filters>' },

    // select file with rev
    { speaker: 'human', text: 'What is the query for <request>files changes by commit 4569455 in aws repo</request>?' },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>select:file type:diff repo:aws rev:4569455</filters>',
    },

    // select removed symbol
    { speaker: 'human', text: 'What is the query for <request>removed usage of toUpperCase</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>toUpperCase</contents><filters>select:commit.diff.removed type:diff</filters>',
    },

    // select removed text with repo
    {
        speaker: 'human',
        text: 'What is the query for <request>removed call to model.fit in deeplearning repo</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>model.fit</contents><filters>select:commit.diff.removed type:diff repo:deeplearning</filters>',
    },

    // select removed text with file
    {
        speaker: 'human',
        text: 'What is the query for <request>removed reference to AuthenticatedUser in graphql/api</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>AuthenticatedUser</contents><filters>select:commit.diff.removed type:diff file:graphql/api</filters>',
    },

    // select removed with who query correction
    { speaker: 'human', text: 'What is the query for <request>who removed <Form /></request>?' },
    {
        speaker: 'assistant',
        text: '<contents><Form /></contents><filters>select:commit.diff.removed type:diff</filters>',
    },

    // select changed with who query correction
    { speaker: 'human', text: 'What is the query for <request>who changed UserAvatar typescript class</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>UserAvatar</contents><filters>type:diff lang:typescript</filters>',
    },

    // select added symbol
    { speaker: 'human', text: 'What is the query for <request>added parseDate</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>parseDate</contents><filters>select:commit.diff.added type:diff</filters>',
    },

    // select added text with repo
    { speaker: 'human', text: 'What is the query for <request>added call to df.sort in nlp repo</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>df.sort</contents><filters>select:commit.diff.added type:diff repo:nlp</filters>',
    },

    // select added text with file
    {
        speaker: 'human',
        text: 'What is the query for <request>added reference to Cookies in src/middleware</request>?',
    },
    {
        speaker: 'assistant',
        text: '<contents>Cookies</contents><filters>select:commit.diff.added type:diff file:src/middleware</filters>',
    },

    // select added with who query correction
    { speaker: 'human', text: 'What is the query for <request>who added UserAvatar</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>UserAvatar</contents><filters>select:commit.diff.added type:diff</filters>',
    },

    // select symbol function
    { speaker: 'human', text: 'What is the query for <request>functions with name like is_.*?$</request>?' },
    { speaker: 'assistant', text: '<contents>is_.*$</contents><filters>select:symbol.function type:symbol</filters>' },

    // select symbol function starts with
    { speaker: 'human', text: 'What is the query for <request>functions starting with Get</request>?' },
    { speaker: 'assistant', text: '<contents>^Get</contents><filters>select:symbol.function type:symbol</filters>' },

    // select symbol function ends with
    { speaker: 'human', text: 'What is the query for <request>functions ending with Delete</request>?' },
    { speaker: 'assistant', text: '<contents>Delete$</contents><filters>select:symbol.function type:symbol</filters>' },

    // select symbol interface
    { speaker: 'human', text: 'What is the query for <request>IProps interfaces</request>?' },
    { speaker: 'assistant', text: '<contents>IProps</contents><filters>select:symbol.interface type:symbol</filters>' },

    // select symbol graphql enums
    { speaker: 'human', text: 'What is the query for <request>enums in graphql schema</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>.*</contents><filters>select:symbol.enum type:symbol file:schema\\.graphql</filters>',
    },

    // select symbol enums in go lang
    { speaker: 'human', text: 'What is the query for <request>go lang enums</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>.*</contents><filters>select:symbol.enum type:symbol lang:golang</filters>',
    },

    // select symbol class with name like
    { speaker: 'human', text: 'What is the query for <request>java class with name like OAuth</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>OAuth</contents><filters>select:symbol.class type:symbol lang:java</filters>',
    },

    // select symbol function with diff correction
    { speaker: 'human', text: 'What is the query for <request>GetUserInfo function changed</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>GetUserInfo</contents><filters>type:diff</filters>',
    },

    // select symbol class with diff and lang correction
    { speaker: 'human', text: 'What is the query for <request>typescript class added</request>?' },
    {
        speaker: 'assistant',
        text: '<contents></contents><filters>select:commit.diff.added type:diff lang:typescript</filters>',
    },

    // select symbol method removed correction
    { speaker: 'human', text: 'What is the query for <request>toString method removed</request>?' },
    {
        speaker: 'assistant',
        text: '<contents>toString</contents><filters>select:commit.diff.removed type:diff</filters>',
    },

    // and operator
    { speaker: 'human', text: 'What is the query for <request>files containing bat and ball</request>?' },
    { speaker: 'assistant', text: '<contents>bat AND ball</contents>' },

    // and operator with lang filter
    { speaker: 'human', text: 'What is the query for <request>python files containing bat and ball</request>?' },
    { speaker: 'assistant', text: '<contents>bat AND ball</contents><filters>lang:python</filters>' },

    // or operator
    { speaker: 'human', text: 'What is the query for <request>fruits or vegetables</request>?' },
    { speaker: 'assistant', text: '<contents>fruits AND vegetables</contents>' },

    // or operator with lang filter
    { speaker: 'human', text: 'What is the query for <request>css files with both media or position</request>?' },
    { speaker: 'assistant', text: '<contents>media OR position</contents><filters>lang:css/</filters>' },

    // fallback
    { speaker: 'human', text: 'What is the query for <request>DFH84fHAg</request>?' },
    {
        speaker: 'assistant',
        text: 'I apologize, but I do not understand the request "DFH84fHAg". Without more context about what is being requested, I cannot generate a valid query.',
    },
    {
        speaker: 'human',
        text: 'NEVER ASK FOR MORE CONTEXT and ALWAYS MAKE A GUESS. If you are unsure, just treat the entire request as a regular expression matching file contents. What is the query for <request>DFH84fHAg</request>?',
    },
    { speaker: 'assistant', text: '<contents>DFH84fHAg</contents><filters></filters>' },

    // actual user query
    { speaker: 'human', text: `What is the query for <request>${input}</request>?` },
    { speaker: 'assistant', text: '' },
]
