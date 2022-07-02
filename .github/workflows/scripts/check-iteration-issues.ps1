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

function Get-UserReference {
    [CmdletBinding()]
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

# PST is the most western timezone we have, i.e. the latest EOD teammates have.
# Therefor the milestone due date refers to the end of PST Friday (which is already Saturday in UTC, and GitHub Actions run in UTC, but this script may also be run locally for testing in any timezone).
$todayInPST = [TimeZoneInfo]::ConvertTimeBySystemTimeZoneId([DateTime]::Today, 'Pacific Standard Time').Date

Write-Information "Current date: $todayInPST"

$milestones = Get-GitHubMilestone -Owner sourcegraph -RepositoryName sourcegraph -State open -Sort DueDate

# Current and next iteration
$relevantMilestones = $milestones |
    Where-Object { $_.Title -like 'Insights iteration*' } |
    Where-Object {
        $iterationStart = $_.DueOn.Date.AddDays(-11) # First Monday of the iteration (due date is always on the second Friday)
        $daysUntilIterationStart = ($iterationStart - $todayInPST).TotalDays

        (
            # Is current iteration?
            ($todayInPST -ge $iterationStart -and $todayInPST -lt $_.DueOn.Date) -or
            # Is next iteration and only a week away (start date less than 7 days away)?
            $daysUntilIterationStart -gt 0 -and $daysUntilIterationStart -lt 7
        )
    } |
    Select-Object -First 2


if (!$relevantMilestones) {
    Write-Warning "No current milestone found for today ($($todayInPST.ToLongDateString()))"
    return
}

$now = Get-Date

# Proposed issues

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $isCurrent = $todayInPST -lt $_.DueOn.Date -and $todayInPST -ge $_.DueOn.Date.AddDays(-11)
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

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


# Estimates

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $isCurrent = $todayInPST -lt $_.DueOn.Date -and $todayInPST -ge $_.DueOn.Date.AddDays(-11)
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

    Write-Information "Milestone: $($milestone.Title) ($label)"

    $itemsWithoutEstimate = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($milestone.Title)`" -assignee:AlicjaSuska" |
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


# Assignees

$parts = $relevantMilestones | ForEach-Object {
    $milestone = $_

    $isCurrent = $todayInPST -lt $_.DueOn.Date -and $todayInPST -ge $_.DueOn.Date.AddDays(-11)
    $label = if ($isCurrent) { 'current' } else { 'next' }
    $viewUrl = $iterationViews[$label]

    Write-Information "Milestone: $($milestone.Title) ($label)"

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
