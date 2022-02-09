# Script to run as a GitHub action on issue and PR updates that will update the associated GitHub Beta project items.

[CmdletBinding()]
param(
    # GitHub GraphQL Node ID of the GitHub Beta project
    [Parameter(Mandatory)]
    [string] $ProjectNodeId,

    # The team/* label to filter issues/PRs by. All issues/PRs that don't have this label will be ignored.
    [Parameter(Mandatory)]
    [string] $TeamLabel,

    # Previously set up webhook URI from https://sourcegraph.slack.com/apps/A0F7XDUAZ
    [Parameter(Mandatory)]
    [string] $SlackWebhookUri,

    # Slack channel to post to
    [Parameter(Mandatory)]
    [string] $SlackChannel
)

# Regex for extracting the "Closes #1234" pattern in GitHub PR descriptions
$fixIssuePattern = "(?:close|fixe?|resolve)(?:[sd])? (?:#|(?<owner>[\w_-]+)/(?<repo>[\w_-]+)#|https://github\.com/(?<owner>[\w_-]+)/(?<repo>[\w_-]+)/issues/)(?<number>\d+)"

switch ($github.event_name) {

    'issues' {
        if (-not ($github.event.issue.labels | Where-Object { $_.name -eq $TeamLabel })) {
            Write-Information "Issue does not have $TeamLabel label, exiting."
            return
        }

        Write-Information "Issue was $($github.event.action)"

        if ($github.event.action -in 'opened', 'labeled', 'milestoned') {
            # If team label was added or issue was just opened, add to project board
            # If added to an iteration, update status and set "proposed by" to the event actor
            # Idempotent, will return the item if already exists in the board (this is fine because we checked for the team label)
            $item = [pscustomobject]$github.event.issue | Add-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId

            if ($item.content.milestone) {
                $proposer = $github.event.sender.login
                Write-Information "Updating issue as 'Proposed for iteration' by @$proposer"

                $item |
                    Set-GitHubBetaProjectItemField -Name 'Status' -Value 'Proposed for iteration' |
                    Set-GitHubBetaProjectItemField -Name 'Proposed by' -Value $proposer


                # Post Slack message

                $size = $item.Fields['Size 🔵']
                $iterationTitle = $item.content.milestone.title
                $issueUrl = $item.content.url

                $stats = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($item.content.milestone.title)`"" |
                    Get-GitHubBetaProjectItem |
                    Where-Object { $_.project.id -eq $ProjectNodeId -and $_.Fields['Status'] -ne 'Done' } |
                    ForEach-Object { $_.Fields['Size 🔵'] ?? 1 } |
                    Measure-Object -AllStats

                $message = "*$proposer* proposed a new issue $($size ? "of *size $size*" : "without a size") for iteration *$($iterationTitle)*:`n" +
                    "$issueUrl`n" +
                    "`n" +
                    "There are now $($stats.Sum) points of open issues in the iteration."

                Write-Information "Sending Slack message:`n$message"

                Send-SlackMessage -Text $message -Username 'Iteration Bot' -IconEmoji ':robot:' -Channel $SlackChannel -Uri $SlackWebhookUri -UnfurlLinks $true
            }
        } else {
            # If issue was closed or reopened, update Status column
            $status = if ($github.event.action -eq 'closed') { 'Done' } else { 'In Progress' }

            [pscustomobject]$github.event.issue |
                # Idempotent, will return the item if already exists
                Add-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId |
                Set-GitHubBetaProjectItemField -FieldName 'Status' -Value $status |
                ForEach-Object { Write-Information "Updated `"Status`" field of project item for $($_.content.url) to `"$status`"" }
        }
    }

    'pull_request' {
        $pr = $github.event.pull_request

        # Ignore merged and closed PRs
        if ($pr.state -ne 'open') {
            return
        }

        $status = if ($pr.draft) { 'In Progress' } else { 'In Review' }

        # Get fixed issues from the PR description
        [regex]::Matches($pr.body, $fixIssuePattern, [Text.RegularExpressions.RegexOptions]::IgnoreCase) |
            ForEach-Object {
                $owner = if ($_.Groups['owner'].Success) { $_.Groups['owner'].Value } else { $github.event.repository.owner.login }
                $repo = if ($_.Groups['repo'].Success) { $_.Groups['repo'].Value } else { $github.event.repository.name }
                $number = $_.Groups['number'].Value
                Write-Information "Found fixed issue $owner/$repo#$number"
                Get-GitHubIssue -Owner $owner -Repository $repo -Number $number
            } |
            Where-Object { $_.labels | Where-Object { $_.name -eq $TeamLabel } } |
            # Idempotent, will return the item if already exists
            Add-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId |
            Set-GitHubBetaProjectItemField -FieldName 'Status' -Value $status |
            ForEach-Object { Write-Information "Updated `"Status`" field of project item for $($_.content.url) to `"$status`"" }
    }
}
