@echo off
REM ============================================================================
REM SPX TileMap Export Script
REM 
REM 自动调用 Godot 导出 TileMap 场景数据到 SPX JSON 格式
REM ============================================================================

setlocal enabledelayedexpansion

REM 配置路径
set GODOT_EXE=C:\Users\tjp\Downloads\Godot_v4.4.1-stable_win64.exe\Godot_v4.4.1-stable_win64.exe
set PROJECT_DIR=D:\projects\spx\tutorial\AITown
set EXPORT_SCRIPT=addons\spx_tilemap_exporter\export_cli.gd
set TARGET_DIR=D:\projects\spx\tutorial\11-Path\assets\tilemaps
set SCENE_NAME=main

REM 检查 Godot 是否存在
if not exist "%GODOT_EXE%" (
    echo ERROR: Godot executable not found at:
    echo   %GODOT_EXE%
    echo.
    echo Please check the path and try again.
    exit /b 1
)

REM 检查项目目录是否存在
if not exist "%PROJECT_DIR%\project.godot" (
    echo ERROR: Godot project not found at:
    echo   %PROJECT_DIR%
    echo.
    echo Please ensure project.godot exists in the directory.
    exit /b 1
)

REM 检查导出脚本是否存在
if not exist "%PROJECT_DIR%\%EXPORT_SCRIPT%" (
    echo ERROR: Export script not found at:
    echo   %PROJECT_DIR%\%EXPORT_SCRIPT%
    echo.
    echo Please ensure the export_cli.gd script exists.
    exit /b 1
)

echo ============================================================================
echo SPX TileMap Export
echo ============================================================================
echo.
echo Godot:   %GODOT_EXE%
echo Project: %PROJECT_DIR%
echo.
echo Starting export...
echo.

REM 运行 Godot 无头模式执行导出脚本
REM 使用临时文件捕获输出，过滤掉 Godot 引擎的 TileSet 内部警告
set TEMP_OUTPUT=%TEMP%\spx_export_output.txt
"%GODOT_EXE%" --headless --path "%PROJECT_DIR%" -s "%EXPORT_SCRIPT%" 2>&1 | findstr /V /C:"Cannot create tile" /C:"TileSetAtlasSource" /C:"has no tile at" /C:"create_tile" /C:"has_alternative_tile" /C:"create_alternative_tile"

set EXIT_CODE=%ERRORLEVEL%

echo.
if %EXIT_CODE% EQU 0 (
    echo ============================================================================
    echo Export completed successfully!
    echo Output saved to: %PROJECT_DIR%\_export\%SCENE_NAME%\
    echo ============================================================================
    echo.
    echo Copying to 11-Path/assets/tilemaps...
    
    REM 定义源目录和目标目录
    set SOURCE_DIR=%PROJECT_DIR%\_export\%SCENE_NAME%
    
    REM 如果目标目录存在，先删除
    if exist "!TARGET_DIR!" (
        echo Removing existing target directory...
        rmdir /s /q "!TARGET_DIR!"
    )
    
    REM 复制源目录到目标位置
    echo Copying files...
    xcopy "!SOURCE_DIR!" "!TARGET_DIR!" /E /I /Q
    
    if !ERRORLEVEL! EQU 0 (
        echo.
        echo ============================================================================
        echo Copy completed successfully!
        echo Copied to: !TARGET_DIR!
        echo ============================================================================
    ) else (
        echo.
        echo ============================================================================
        echo WARNING: Copy failed with error code: !ERRORLEVEL!
        echo ============================================================================
    )
) else (
    echo ============================================================================
    echo Export failed with error code: %EXIT_CODE%
    echo ============================================================================
)

endlocal
exit /b %EXIT_CODE%
