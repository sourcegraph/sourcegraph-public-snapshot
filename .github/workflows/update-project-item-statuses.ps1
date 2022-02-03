# Script to run as a GitHub action on issue and PR updates that will update the associated GitHub Beta project items.

[CmdletBinding()]
param(
    # GitHub GraphQL Node ID of the GitHub Beta project
    [Parameter(Mandatory)]
    [string] $ProjectNodeId,

    # The team/* label to filter issues/PRs by. All issues/PRs that don't have this label will be ignored.
    [Parameter(Mandatory)]
    [string] $TeamLabel
)

# Regex for extracting the "Closes #1234" pattern in GitHub PR descriptions
$fixIssuePattern = "(?:close|fixe?|resolve)(?:[sd])? (?:#|(?<owner>[\w_-]+)/(?<repo>[\w_-]+)#|https://github\.com/(?<owner>[\w_-]+)/(?<repo>[\w_-]+)/issues/)(?<number>\d+)"

switch ($github.event_name) {

    'issues' {
        # Find project item for the issue
        # THIS DOES NOT SCALE AS THE PROJECT GETS LARGE, but it's the only way possible afaict.
        # One way this is mitigated is that this request is streamed/paginated in order of most-recent-first,
        # which means we can hope the issue is usually found in the first page(s).

        if (-not ($github.event.issue.labels | Where-Object { $_.name -eq $TeamLabel })) {
            Write-Information "Issue does not have $TeamLabel label, exiting."
            return
        }

        $status = if ($github.event.action -eq 'closed') { 'Done' } else { 'In Progress' }

        Get-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId |
            Where-Object { $_.content -and $_.content.number -eq $github.event.issue.number } |
            Select-Object -First 1 |
            Set-GitHubBetaProjectItemField -FieldName 'Status' -Value $status |
            ForEach-Object { Write-Information "Updated `"Status`" field of project item for $($_.content.url) to `"$status`"" }
    }

    'pull_request' {
        $pr = $github.event.pull_request

        # Ignore merged and closed PRs
        if ($pr.state -ne 'open') {
            return
        }

        $status = if ($pr.draft) { 'In Progress' } else { 'In Review' }

        # Get fixed issues from the PR description
        $fixedIssues = [regex]::Matches($pr.body, $fixIssuePattern, [Text.RegularExpressions.RegexOptions]::IgnoreCase) |
            ForEach-Object {
                $owner = if ($_.Groups['owner'].Success) { $_.Groups['owner'].Value } else { $github.event.repository.owner.login }
                $repo = if ($_.Groups['repo'].Success) { $_.Groups['repo'].Value } else { $github.event.repository.name }
                $number = $_.Groups['number'].Value
                Get-GitHubIssue -Owner $owner -Repository $repo -Number $number
            } |
            Where-Object { $_.labels | Where-Object { $_.name -eq $TeamLabel } }

        if (!$fixedIssues) {
            Write-Information "No fixed issues with $TeamLabel label referenced from PR description, exiting."
            return
        }

        Write-Information "Fixed issues:"
        $fixedIssues | ForEach-Object HtmlUrl | Write-Information

        # Find project items for the issues the PR references
        Get-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId |
            Where-Object {
                $item = $_
                $fixedIssues | Where-Object {
                    $item.content -and
                    $_.Number -eq $item.content.number -and
                    $_.Owner -eq $item.content.repository.owner.login -and
                    $_.RepositoryName -eq $item.content.repository.name
                }
            } |
            Select-Object -First $fixedIssues.Count |
            Set-GitHubBetaProjectItemField -FieldName 'Status' -Value $status |
            ForEach-Object { Write-Information "Updated `"Status`" field of project item for $($_.content.url) to `"$status`"" }
    }
}
