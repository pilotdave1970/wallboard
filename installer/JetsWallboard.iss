; Jet's Wallboard installer (Inno Setup 6, built on Windows by GitHub Actions).
; One installer, no choices to make: installs for the current user (no admin prompt) into
; %LOCALAPPDATA%\Programs\Jets Wallboard and adds Start menu and desktop shortcuts.
; Startup, full screen and other options live in the board: Settings > Board > This PC.
;
; The board's update button runs this same installer silently with /update=1 /browserpid=N,
; and the new version takes over the board window that's already open.
;
; Build:  iscc /DMyAppVersion=1.1.0 /DSourceExe=C:\path\JetsWallboard.exe installer\JetsWallboard.iss

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
DisableReadyPage=yes
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

[Files]
Source: "{#SourceExe}"; DestDir: "{app}"; DestName: "{#MyAppExe}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExe}"

[Run]
; After a normal install: offer to open it.
Filename: "{app}\{#MyAppExe}"; Description: "Open Jet's Wallboard now"; Flags: nowait postinstall skipifsilent
; After an update from inside the board: start the new version on the window that's already open.
Filename: "{app}\{#MyAppExe}"; Parameters: "--browser-pid={param:browserpid|-1}"; Flags: nowait; Check: IsUpdate

[InstallDelete]
; leftovers from older versions that updated themselves in place
Type: files; Name: "{app}\{#MyAppExe}.old"
Type: files; Name: "{app}\{#MyAppExe}.new"

[UninstallRun]
; Stop starting with Windows.
Filename: "{sys}\reg.exe"; Parameters: "delete HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v JetsWallboard /f"; Flags: runhidden; RunOnceId: "RemoveStartup"

[UninstallDelete]
Type: files; Name: "{app}\{#MyAppExe}.old"
Type: files; Name: "{app}\{#MyAppExe}.new"
; Settings and pictures in %APPDATA%\JetsWallboard are kept on purpose.

[Code]
function IsUpdate(): Boolean;
begin
  Result := ExpandConstant('{param:update|0}') = '1';
end;

// Stop a running board helper before installing or uninstalling, so the .exe can be replaced.
procedure StopWallboard();
var
  Code: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM {#MyAppExe}', '', SW_HIDE, ewWaitUntilTerminated, Code);
  Sleep(500);
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
