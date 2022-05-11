# Script to run as a GitHub action on issue and PR updates that will update the associated GitHub Beta project items.

[CmdletBinding()]
param(
    # GitHub GraphQL Node ID of the GitHub Beta project
    [Parameter(Mandatory)]
    [string] $ProjectNodeId,

    # The team/* label to filter issues/PRs by. `labeled` events for other labels than this label will be ignored.
    [Parameter(Mandatory)]
    [string] $TeamLabel,

    # A regular expression pattern to match milestone titles by.
    # All `milestoned` events not matching the pattern will be ignored.
    [Parameter(Mandatory)]
    [regex] $TeamIterationMilestonePattern,

    # Previously set up webhook URI from https://sourcegraph.slack.com/apps/A0F7XDUAZ
    [Parameter(Mandatory)]
    [string] $SlackWebhookUri,

    # Slack channel to post to
    [Parameter(Mandatory)]
    [string] $SlackChannel
)

switch ($github.event_name) {

    'issues' {
        Write-Information "Issue was $($github.event.action)"

        if ($github.event.action -eq 'labeled') {
            # If issue was labeled, make sure to only consider the team label being added.
            if ($github.event.label.name -ne $TeamLabel) {
                Write-Information "Labeled with non-team label ($($github.event.label.name)), exiting."
                return
            }

            # Add issue to the team project board. Idempotent.
            [pscustomobject]$github.event.issue | Add-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId

        } elseif ($github.event.action -eq 'milestoned') {
            # If issue was milestoned, make sure to only consider team iteration milestones
            if ($github.event.issue.milestone.title -notmatch $TeamIterationMilestonePattern) {
                Write-Information "Milestoned with non-matching milestone ($($github.event.issue.milestone.title)), exiting. Pattern: $TeamIterationMilestonePattern"
                return
            }

            # Add issue to the team project board. Idempotent, will return the item if already exists in the board.
            $item = [pscustomobject]$github.event.issue | Add-GitHubBetaProjectItem -ProjectNodeId $ProjectNodeId

            # Update status and set "Proposed by" to the event actor

            $proposer = $github.event.sender.login
            Write-Information "Updating issue as 'Proposed for iteration' by @$proposer"

            if ($item.Fields['Status'] -notin 'In Progress', 'In Review', 'Done') {
                $item |
                    Set-GitHubBetaProjectItemField -Name 'Status' -Value 'Proposed for iteration' |
                    Set-GitHubBetaProjectItemField -Name 'Proposed by' -Value $proposer
            }

            # Post Slack message

            $stats = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($item.content.milestone.title)`"" |
                Get-GitHubBetaProjectItem |
                Where-Object { $_.project.id -eq $ProjectNodeId -and $_.Fields['Status'] -ne 'Done' } |
                ForEach-Object { $_.Fields['Size 🔵'] ?? 1 } |
                Measure-Object -AllStats

            $color = if ($item.content.state -eq 'OPEN') { '#1A7F37' } else { '#8250DF' }

            $message = "*$proposer* proposed a new issue for iteration <$($item.content.milestone.url)|$($item.content.milestone.title)>.`n" +
                "There are now $($stats.Sum) points of open issues in the iteration."

            # Plain text fallback for contexts without formatting capability, e.g. push notifications
            $fallback = "*$proposer* proposed a new issue for iteration $($item.content.milestone.title): #$($item.content.number) $($item.content.title). There are now $($stats.Sum) points of open issues in the iteration."

            New-SlackMessageAttachment `
                -Pretext $message `
                -Color $color `
                -AuthorName $item.content.author.login `
                -AuthorIcon $item.content.author.avatarUrl `
                -Title "#$($item.content.number) $($item.content.title)" `
                -TitleLink $item.content.url `
                -Text $item.content.bodyText.Substring(0, [System.Math]::Min(1000, $item.content.bodyText.Length)) `
                -Fields @(
                    @{ title = 'Size'; value = $item.Fields['Size 🔵']; short = $true },
                    @{ title = 'Importance'; value = $item.Fields['Importance']; short = $true },
                    @{ title = 'Labels'; value = $item.content.labels | ForEach-Object name | Join-String -Separator ', '; short = $true },
                    @{ title = 'Assignee'; value = $item.content.assignees | ForEach-Object login | Join-String -Separator ', '; short = $true }
                ) `
                -Fallback $fallback |
                New-SlackMessage -Username 'Iteration Bot' -IconEmoji ':robot:' -Channel $SlackChannel |
                Send-SlackMessage -Uri $SlackWebhookUri

        } else {
            if (-not ($github.event.issue.labels | Where-Object { $_.name -eq  $TeamLabel })) {
                Write-Information "Closed / re-opened issue labeled without team label $TeamLabel, exiting."
                return
            }

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

        # Regex for extracting the "Closes #1234" pattern in GitHub PR descriptions
        $fixIssuePattern = "(?:close|fixe?|resolve)(?:[sd])? (?:#|(?<owner>[\w_-]+)/(?<repo>[\w_-]+)#|https://github\.com/(?<owner>[\w_-]+)/(?<repo>[\w_-]+)/issues/)(?<number>\d+)"

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
