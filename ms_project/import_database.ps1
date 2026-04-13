# 数据库导入脚本
# 将导出的SQL文件导入到MySQL容器中
# 使用前提：MySQL容器正在运行，且配置与导出时一致

param(
    [string]$sqlFile = "msproject_full_dump.sql",
    [string]$dbName = "msproject",
    [string]$mysqlUser = "root",
    [string]$mysqlPassword = "root"
)

Write-Host "开始导入数据库..." -ForegroundColor Green
Write-Host "SQL文件: $sqlFile" -ForegroundColor Yellow
Write-Host "数据库名: $dbName" -ForegroundColor Yellow

# 检查文件是否存在
if (-not (Test-Path $sqlFile)) {
    Write-Host "错误：SQL文件不存在: $sqlFile" -ForegroundColor Red
    exit 1
}

# 检查MySQL容器是否运行
$mysqlRunning = docker-compose ps mysql --quiet
if (-not $mysqlRunning) {
    Write-Host "MySQL容器未运行，尝试启动..." -ForegroundColor Yellow
    docker-compose up -d mysql
    # 等待MySQL初始化
    Write-Host "等待MySQL初始化..." -ForegroundColor Yellow
    Start-Sleep -Seconds 15
}

# 检查文件大小
$fileSize = (Get-Item $sqlFile).Length
Write-Host "文件大小: $($fileSize / 1KB) KB" -ForegroundColor Cyan

# 导入数据库
Write-Host "正在导入数据库，请稍候..." -ForegroundColor Green

try {
    Get-Content $sqlFile | docker-compose exec -T mysql mysql -u$mysqlUser -p$mysqlPassword $dbName
    Write-Host "数据库导入成功！" -ForegroundColor Green
    
    # 验证导入结果
    Write-Host "验证导入结果..." -ForegroundColor Yellow
    docker-compose exec -T mysql mysql -u$mysqlUser -p$mysqlPassword $dbName -e "SHOW TABLES;" | Select-Object -Skip 1
    
    Write-Host "数据库 '$dbName' 已成功从 '$sqlFile' 导入。" -ForegroundColor Green
}
catch {
    Write-Host "导入失败: $_" -ForegroundColor Red
    exit 1
}