param (
    [string]$ChangeId = "I834a245f927b1432db41d12a3b2bd3e9718d009a",
    [string]$Patchset = "1",
    [bool]$EnableContext = $true,
    [bool]$React = $false,
    [bool]$AutoPublish = $true,
    [string]$Url = "http://localhost:8000/reviews/run"
)

$Body = @{
    changeId      = $ChangeId
    patchset      = $Patchset
    enableContext = $EnableContext
    react         = $React
    autoPublish   = $AutoPublish
} | ConvertTo-Json

Write-Host "Triggering review for ChangeId: $ChangeId, Patchset: $Patchset"
try {
    $response = Invoke-RestMethod -Uri $Url -Method Post -Body $Body -ContentType "application/json"
    Write-Host "Response:"
    $response | ConvertTo-Json -Depth 5
}
catch {
    Write-Error "Failed to trigger review: $_"
}
