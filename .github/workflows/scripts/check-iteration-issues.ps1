# Checks the issues in the current and next iteration for
# - "Proposed" issues that have been sitting in proposed for longer than 24h
# - Issues without estimate

[CmdletBinding()]
param(
    # GitHub GraphQL Node ID of the GitHub Beta project
    [Parameter(Mandatory)]
    [string] $ProjectNodeId,

    # Previously set up webhook URI from https://sourcegraph.slack.com/apps/A0F7XDUAZ
    [Parameter(Mandatory)]
    [string] $SlackWebhookUri,

    # Slack channel to post to
    [Parameter(Mandatory)]
    [string] $SlackChannel
)

# Map from GitHub username to Slack user ID
$slackUserIds = @{
    'AlicjaSuska' = 'U0166SK4BPC'
    'chwarwick' = 'U035Z2VHWTH'
    'coury-clark' = 'U021CRVEQ5V'
    'CristinaBirkel' = 'U02CE9E6N87'
    'felixfbecker' = 'U2LRFURMW'
    'Joelkw' = 'U01ADU7BKML'
    'leonore' = 'U034ZR3DMKK'
    'unclejustin' = 'U029KKZV916'
    'vovakulikov' = 'U01SD823C9W'
}

$designerGitHubHandle = 'AlicjaSuska'

function Get-UserReference {
    [CmdletBinding()]
    [OutputType([string])]
    param(
        [Parameter(Mandatory, ValueFromPipeline, ValueFromPipelineByPropertyName)]
        [string] $Login
    )
    process {
        $slackUserId = $slackUserIds[$Login]
        if ($slackUserId) {
            "<@$slackUserId>"
        } else {
            "<https://github.com/$Login|$Login>"
        }
    }
}

$iterationViews = @{
    current = 'https://github.com/orgs/sourcegraph/projects/200/views/1'
    next = 'https://github.com/orgs/sourcegraph/projects/200/views/4'
}

$slackParams = @{
    Channel = $SlackChannel
    Username = 'Iteration Bot'
    IconEmoji = ':robot:'
    LinkNames = $true
    Uri = $SlackWebhookUri
}

$now = Get-Date
Write-Information "Current date: $now"

$milestones = Get-GitHubMilestone -Owner sourcegraph -RepositoryName sourcegraph -State open -Sort DueDate

# Current and next iteration
$relevantMilestones = $milestones |
    Where-Object { $_.Title -like 'Insights iteration*' } |
    Where-Object {
        $iterationStart = $_.DueOn.Date.AddDays(-11) # First Monday of the iteration (due date is always on the second Friday)
        $daysUntilIterationStart = ($iterationStart - $now).TotalDays

        (
            # Is current iteration?
            ($now -ge $iterationStart -and $now -lt $_.DueOn.Date) -or
            # Is next iteration and only a week away (start date less than 7 days away)?
            $daysUntilIterationStart -gt 0 -and $daysUntilIterationStart -lt 7
        )
    } |
    Select-Object -First 2


if (!$relevantMilestones) {
    Write-Warning "No current milestone found for today ($($now.ToLongDateString()))"
    return
}

# Estimates

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $iterationStart = $_.DueOn.Date.AddDays(-11)
    $daysUntilIterationStart = ($iterationStart - $now).TotalDays
    $isCurrent = $now -lt $_.DueOn.Date -and $now -ge $iterationStart
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

    # Only start posting after Tuesday before the next iteration, or during the current.
    # The estimations are important for the PM's pre-planning/proposals before our Thursday planning meeting.
    if (-not ($isCurrent -or $daysUntilIterationStart -lt 6)) {
        return
    }

    Write-Information "Milestone: $($milestone.Title) ($label)"

    $itemsWithoutEstimate = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($milestone.Title)`" -assignee:$designerGitHubHandle" |
        Get-GitHubBetaProjectItem |
        Where-Object { $_.project.id -eq $ProjectNodeId -and !$_.Fields['Size 🔵'] }

    if (!$itemsWithoutEstimate) {
        Write-Information "No items without estimate"
        return
    }

    $list = $itemsWithoutEstimate | ForEach-Object {
        $assignees = ($_.content.assignees | Get-UserReference) -join ', '
        $icon = if ($_.content.state -eq 'open') { ':issue:' } else { ':issueclosed:' }
        "$icon *<$($_.content.url)|$($_.content.title)>* created by $($_.content.author | Get-UserReference), $($_.content.assignees ? "assigned to $assignees" : "unassigned")"
    }

    "There are $($itemsWithoutEstimate.Count) issues in *<$viewUrl|$($milestone.Title)>* (*$label iteration*) that have *no estimate*:`n`n$($list -join "`n")"
}

if ($parts) {
    $message = "$($parts -join "`n`n")`n`nAuthors or assignees, can you give these an estimate?"
    Write-Information "Sending Slack message:`n$message"
    Send-SlackMessage @slackParams -Text $message
}

# Proposed issues

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $iterationStart = $_.DueOn.Date.AddDays(-11)
    $daysUntilIterationStart = ($iterationStart - $now).TotalDays
    $isCurrent = $now -lt $_.DueOn.Date -and $now -ge $iterationStart
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

    # Only start posting after Thursday before the next iteration, or during the current.
    # After Thursday, all items should be in Todo or moved out of the iteration.
    if (-not ($isCurrent -or $daysUntilIterationStart -lt 4)) {
        return
    }

    Write-Information "Milestone: $($milestone.Title) ($label)"

    $proposedItems = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($milestone.Title)`"" |
        Where-Object {
            $milestoned = $_ | Get-GitHubIssueTimeline | Where-Object { $_.Event -eq 'milestoned' } | Select-Object -Last 1
            ($now - $milestoned.CreatedAt).TotalHours -ge 24
        } |
        Get-GitHubBetaProjectItem |
        Where-Object { $_.project.id -eq $ProjectNodeId -and $_.Fields['Status'] -eq 'Proposed for iteration' }

    if (!$proposedItems) {
        Write-Information 'No items in "Proposed"'
        return
    }

    $list = $proposedItems | ForEach-Object {
        $icon = if ($_.content.state -eq 'open') { ':issue:' } else { ':issueclosed:' }
        $size = if ($_.Fields['Size 🔵']) { "size *$($_.Fields['Size 🔵'])*" } else { '_unestimated_' }
        "$icon *<$($_.content.url)|$($_.content.title)>* ($size) proposed by $(Get-UserReference -Login $_.Fields['Proposed by'])"
    }

    "There are $($proposedItems.Count) issues that have been in *Proposed* for *<$viewUrl|$($milestone.Title)>* (*$label iteration*) for longer than 24h:`n`n$($list -join "`n")"
}

if ($parts) {
    $message = "$($parts -join "`n`n")`n`nProposers and @joel, can you help drive these to a resolution (whether they should go to _Todo_ or not)?"
    Write-Information "Sending Slack message:`n$message"
    Send-SlackMessage @slackParams -Text $message
}


# Assignees

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $isCurrent = $now -lt $_.DueOn.Date -and $now -ge $_.DueOn.Date.AddDays(-11)
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

    Write-Information "Milestone: $($milestone.Title) ($label)"

    # Only start posting after Thursday before the next iteration, or during the current.
    # After Thursday, all items should have assignees.
    if (-not ($isCurrent -or $daysUntilIterationStart -lt 4)) {
        return
    }

    $itemsWithoutAssignee = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($milestone.Title)`" no:assignee" |
        Get-GitHubBetaProjectItem |
        Where-Object { $_.project.id -eq $ProjectNodeId -and $_.Fields['Status'] -ne 'Proposed for iteration' } # It's okay when issues are not assigned yet when they are proposed

    if (!$itemsWithoutAssignee) {
        Write-Information "No items without assignee"
        return
    }

    $list = $itemsWithoutAssignee | ForEach-Object {
        $icon = if ($_.content.state -eq 'open') { ':issue:' } else { ':issueclosed:' }
        $size = if ($_.Fields['Size 🔵']) { "size *$($_.Fields['Size 🔵'])*" } else { '_unestimated_' }
        $labels = ($_.content.labels | ForEach-Object { "``$($_.name)``" }) -join ', '
        "$icon *<$($_.content.url)|$($_.content.title)>* ($size) $labels"
    }

    "There are $($itemsWithoutAssignee.Count) issues planned for *<$viewUrl|$($milestone.Title)>* (*$label iteration*) that still *need an assignee*:`n`n$($list -join "`n")"
}

if ($parts) {
    $message = "$($parts -join "`n`n")`n`nWho makes the most sense to work on these?"
    Write-Information "Sending Slack message:`n$message"
    Send-SlackMessage @slackParams -Text $message
}
