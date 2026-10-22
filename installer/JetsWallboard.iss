; Jet's Wallboard installer (Inno Setup 6, built on Windows by GitHub Actions).
; Installs for the current user only (no admin prompt) into
; %LOCALAPPDATA%\Programs\Jets Wallboard, so the app can update itself later.
;
; Build:  iscc /DMyAppVersion=1.1.0 /DSourceExe=..\JetsWallboard.exe installer\JetsWallboard.iss

#ifndef MyAppVersion
  #define MyAppVersion "0.0.0"
#endif
#ifndef SourceExe
  #define SourceExe "..\JetsWallboard.exe"
#endif

#define MyAppName "Jet's Wallboard"
#define MyAppExe "JetsWallboard.exe"
#define MyAppURL "https://github.com/pilotdave1970/wallboard"

[Setup]
AppId={{8F3C2A71-5B4E-4C1D-9A6F-2E7B0D4C9A15}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher=pilotdave1970
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={localappdata}\Programs\Jets Wallboard
DisableDirPage=yes
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=..\dist
OutputBaseFilename=JetsWallboardSetup
SetupIconFile=..\assets\icon.ico
UninstallDisplayIcon={app}\{#MyAppExe}
UninstallDisplayName={#MyAppName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
CloseApplications=yes
RestartApplications=no

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Shortcuts:"
Name: "fullscreenicon"; Description: "Create a full-screen shortcut on the desktop"; GroupDescription: "Shortcuts:"; Flags: unchecked
Name: "startup"; Description: "Start the wallboard when Windows starts"; GroupDescription: "Startup:"; Flags: unchecked
Name: "startup\full"; Description: "Full screen"; Flags: unchecked

[Files]
Source: "{#SourceExe}"; DestDir: "{app}"; DestName: "{#MyAppExe}"; Flags: ignoreversion
Source: "..\package\README.txt"; DestDir: "{app}"; Flags: ignoreversion isreadme

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"
Name: "{autoprograms}\{#MyAppName} (Full Screen)"; Filename: "{app}\{#MyAppExe}"; Parameters: "--fullscreen"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"; Tasks: desktopicon
Name: "{autodesktop}\{#MyAppName} (Full Screen)"; Filename: "{app}\{#MyAppExe}"; Parameters: "--fullscreen"; Tasks: fullscreenicon
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"; Tasks: startup and not startup\full
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"; Parameters: "--fullscreen"; Tasks: startup\full

[Run]
Filename: "{app}\{#MyAppExe}"; Description: "Open Jet's Wallboard now"; Flags: nowait postinstall skipifsilent

[InstallDelete]
; leftovers from a self-update
Type: files; Name: "{app}\{#MyAppExe}.old"
Type: files; Name: "{app}\{#MyAppExe}.new"

[UninstallDelete]
Type: files; Name: "{app}\{#MyAppExe}.old"
Type: files; Name: "{app}\{#MyAppExe}.new"
; Settings and pictures in %APPDATA%\JetsWallboard are kept on purpose.

[Code]
// Stop a running board (and its helper) before installing or uninstalling, so the .exe can be replaced.
procedure StopWallboard();
var
  Code: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM {#MyAppExe}', '', SW_HIDE, ewWaitUntilTerminated, Code);
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  StopWallboard();
  Result := '';
end;

function InitializeUninstall(): Boolean;
begin
  StopWallboard();
  Result := True;
end;
