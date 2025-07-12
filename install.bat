@echo off
echo Installing Solana Wallet Token Tracker...

rem Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Go is not installed. Please install Go 1.18 or newer.
    exit /b 1
)

rem Download dependencies
echo Downloading dependencies...
go mod tidy

rem Build the application
echo Building application...
go build -o tracker.exe ./cmd/tracker

rem Create default configuration if not exists
if not exist config.json (
    echo Creating default configuration file...
    copy config.example.json config.json
    echo Please edit config.json with your Solana wallet addresses and token mint addresses.
)

echo Installation complete! Run tracker.exe to start the application.