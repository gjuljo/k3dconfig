@SETLOCAL EnableDelayedExpansion
@echo off
for /f "skip=1 tokens=1-6 delims= " %%a in ('wmic path Win32_LocalTime Get Day^,Hour^,Minute^,Month^,Second^,Year /Format:table') do (
    IF NOT "%%~f"=="" (
        set /a FormattedDate=10000 * %%f + 100 * %%d + %%a
        set FormattedDate=!FormattedDate:~0,4!-!FormattedDate:~-4,2!-!FormattedDate:~-2,2!
        set /a FormattedTime=%%b * 10000 + %%c * 100 + %%e
        set FormattedTime=!FormattedTime:~0,2!:!FormattedTime:~-4,2!:!FormattedTime:~-2,2!
    )
)
set FULLDATE=!FormattedDate!T!FormattedTime!Z
go build -ldflags="-w -s -X main.VERSION=LOCAL -X main.BUILDDATE=%FULLDATE%"