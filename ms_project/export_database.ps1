# 数据库导出脚本
# 导出完整的数据库结构和数据，便于在其他环境部署
# 生成标准SQL文件，包含表结构、数据、存储过程、触发器和事件

param(
    [string]$outputFile = "",  # 输出文件名，为空则使用时间戳生成
    [string]$dbName = "msproject",
    [string]$mysqlUser = "root",
    [string]$mysqlPassword = "root",
    [switch]$compress = $false  # 是否压缩输出
)

# 设置输出文件名
if ([string]::IsNullOrEmpty($outputFile)) {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $outputFile = "${dbName}_full_dump_${timestamp}.sql"
    if ($compress) {
        $outputFile = "${dbName}_full_dump_${timestamp}.sql.gz"
    }
}

Write-Host "开始导出数据库..." -ForegroundColor Green
Write-Host "数据库: $dbName" -ForegroundColor Yellow
Write-Host "输出文件: $outputFile" -ForegroundColor Yellow

# 检查MySQL容器是否运行
$mysqlRunning = docker-compose ps mysql --quiet
if (-not $mysqlRunning) {
    Write-Host "MySQL容器未运行，尝试启动..." -ForegroundColor Yellow
    docker-compose up -d mysql
    # 等待MySQL初始化
    Write-Host "等待MySQL初始化..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
}

# 构建mysqldump命令
$dumpCommand = "docker-compose exec -T mysql mysqldump -u$mysqlUser -p$mysqlPassword $dbName"
$dumpOptions = @(
    "--routines",           # 包含存储过程和函数
    "--triggers",           # 包含触发器
    "--events",             # 包含事件
    "--complete-insert",    # 使用完整的INSERT语句
    "--extended-insert",    # 使用多值INSERT语法
    "--single-transaction", # 在单个事务中导出，确保一致性
    "--default-character-set=utf8mb4",  # 设置字符集
    "--skip-lock-tables"    # 跳过锁表（配合--single-transaction使用）
)

$fullCommand = "$dumpCommand $($dumpOptions -join ' ')"

Write-Host "执行导出命令..." -ForegroundColor Cyan
Write-Host "命令: $fullCommand" -ForegroundColor Gray

# 执行导出
try {
    if ($compress) {
        # 压缩输出
        Invoke-Expression "$fullCommand | gzip -9 > $outputFile"
    } else {
        # 直接输出到文件
        Invoke-Expression "$fullCommand > $outputFile"
    }
    
    # 检查导出结果
    if (Test-Path $outputFile) {
        $fileSize = (Get-Item $outputFile).Length
        $fileSizeKB = [math]::Round($fileSize / 1KB, 2)
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        
        Write-Host "导出成功！" -ForegroundColor Green
        Write-Host "文件: $outputFile" -ForegroundColor Cyan
        Write-Host "大小: $fileSize 字节 ($fileSizeKB KB, $fileSizeMB MB)" -ForegroundColor Cyan
        
        # 验证文件内容
        Write-Host "验证导出文件..." -ForegroundColor Yellow
        
        if (-not $compress) {
            # 统计表数量
            $tableCount = (Select-String -Path $outputFile -Pattern 'CREATE TABLE' | Measure-Object).Count
            $insertCount = (Select-String -Path $outputFile -Pattern 'INSERT INTO' | Measure-Object).Count
            
            Write-Host "包含表数量: $tableCount" -ForegroundColor Cyan
            Write-Host "数据插入语句: $insertCount" -ForegroundColor Cyan
            
            # 检查文件结尾
            $lastLine = Get-Content $outputFile -Tail 1
            if ($lastLine -match 'Dump completed') {
                Write-Host "转储完成标记: $lastLine" -ForegroundColor Green
            }
        }
        
        # 列出包含的表
        Write-Host "`n数据库 '$dbName' 已成功导出到 '$outputFile'" -ForegroundColor Green
        Write-Host "使用 './import_database.ps1 -sqlFile $outputFile' 导入到其他环境" -ForegroundColor Yellow
    } else {
        Write-Host "错误：导出文件未创建" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "导出失败: $_" -ForegroundColor Red
    exit 1
}

# 显示使用说明
Write-Host "`n=== 使用说明 ===" -ForegroundColor Magenta
Write-Host "1. 将 $outputFile 复制到目标服务器"
Write-Host "2. 确保目标服务器有相同的 docker-compose 配置"
Write-Host "3. 运行导入脚本: ./import_database.ps1 -sqlFile $outputFile"
Write-Host "4. 或手动导入: Get-Content $outputFile | docker-compose exec -T mysql mysql -uroot -proot msproject"