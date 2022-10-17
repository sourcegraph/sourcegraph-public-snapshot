# Gathers stats on issue sizes in a GitHub iteration board for an iteration ending today and posts a summary to Slack

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

# PST is the most western timezone we have, i.e. the latest EOD teammates have.
# Therefor the milestone due date refers to the end of PST Friday (which is already Saturday in UTC, and GitHub Actions run in UTC, but this script may also be run locally for testing in any timezone).
$todayInPST = [TimeZoneInfo]::ConvertTimeBySystemTimeZoneId([DateTime]::Today, 'Pacific Standard Time').Date

$currentMilestone = Get-GitHubMilestone -Owner sourcegraph -RepositoryName sourcegraph -State open |
    Where-Object { $_.Title -like 'Insights iteration*' -and $todayInPST -le $_.DueOn.Date -and $todayInPST -ge $_.DueOn.Date.AddDays(-11) }

if (!$currentMilestone) {
    Write-Warning "No current milestone found for today ($($todayInPST.ToLongDateString()))"
    return
}

Write-Information "Current milestone for today: $($currentMilestone.Title)"

$currentIterationItems = Find-GitHubIssue "org:sourcegraph is:issue milestone:`"$($currentMilestone.Title)`"" |
    Get-GitHubBetaProjectItem |
    Where-Object { $_.project.id -eq $ProjectNodeId }

$finishedItems = $currentIterationItems | Where-Object { $_.Fields['Status'] -eq 'Done' }
$notSized = $currentIterationItems | Where-Object { !$_.Fields['Size 🔵'] }

# Maps an input item to its size field, assuming 1 for unsized issues.
filter Get-IterationItemSize {
    $_.Fields['Size 🔵'] ?? 1
}

# Tests whether the given project item has a given GitHub label and returns a boolean
function Test-ItemHasLabel {
    [OutputType([bool])]
    param(
        [Parameter(Mandatory)] [psobject] $Item,
        [Parameter(Mandatory)] [string] $Label
    )

    [bool]($Item.Content.Labels | Where-Object { $_.Name -eq $Label })
}

$stats = $finishedItems | Get-IterationItemSize | Measure-Object -AllStats
$frontendStats = $finishedItems | Where-Object { Test-ItemHasLabel -Item $_ -Label 'webapp' } | Get-IterationItemSize | Measure-Object -AllStats
$backendStats = $finishedItems | Where-Object { Test-ItemHasLabel -Item $_ -Label 'backend' } | Get-IterationItemSize | Measure-Object -AllStats
$notFinished = $currentIterationItems | Where-Object { $_.Fields['Status'] -ne 'Done' } | Get-IterationItemSize | Measure-Object -AllStats

$message = "
Beep bop, this is your friendly iteration bot, with some fresh stats to help with our next iteration planning! :spiral_calendar_pad:

*$($currentMilestone.Title) (current)*
Sum of finished issues: :large_blue_circle: *$($stats.Sum)* | :desktop_computer: Frontend: $($frontendStats.Sum) | :database: Backend: $($backendStats.Sum)
_$($stats.Count) issues, average size $($stats.Average.ToString('#.##')), smallest $($stats.Minimum), largest $($stats.Maximum)_

:issue: Not finished: :large_yellow_circle: $($notFinished.Sum) ($($notFinished.Count) issues)
:grey_question: Not sized: $($notSized.Count) issues
"

Write-Information "Sending Slack message:`n$message"

Send-SlackMessage -Uri $SlackWebhookUri -Channel $SlackChannel -Username 'Iteration Bot' -IconEmoji ':robot:' -Parse full -Text $message
