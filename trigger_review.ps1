param (
    [string]$ChangeNum = "59",
    [string]$Patchset = "1",
    [bool]$EnableContext = $true,
    [bool]$React = $false,
    [bool]$AutoPublish = $true,
    [string]$Url = "http://localhost:8000/reviews/run"
)

$Body = @{
    changeNum     = $ChangeNum
    patchset      = $Patchset
    enableContext = $EnableContext
    react         = $React
    autoPublish   = $AutoPublish
} | ConvertTo-Json

Write-Host "Triggering review for ChangeNum: $ChangeNum, Patchset: $Patchset"
try {
    $response = Invoke-RestMethod -Uri $Url -Method Post -Body $Body -ContentType "application/json"
    Write-Host "Response:"
    $response | ConvertTo-Json -Depth 5
}
catch {
    Write-Error "Failed to trigger review: $_"
}
