[Setup]
AppName=Stock Opname Application
AppVersion={{.Version}}
AppPublisher=Ginfreakz
AppPublisherURL=https://github.com/ginfreakz/stock-opname-app
DefaultDirName={autopf}\Stock Opname App
DefaultGroupName=Stock Opname App
UninstallDisplayIcon={app}\stock-opname-app.exe
Compression=lzma2
SolidCompression=yes
OutputDir=installer
OutputBaseFilename=stock-opname-app_{{.Version}}_windows_amd64
; "ArchitecturesAllowed=x64" specifies that this installer is only for x64 systems.
ArchitecturesAllowed=x64
; "ArchitecturesInstallIn64BitMode=x64" requests that the install be
; done in "64-bit mode" on x64 systems, meaning it should use the
; native 64-bit Program Files directory and Registry entries.
ArchitecturesInstallIn64BitMode=x64
DisableDirPage=no
LicenseFile=LICENSE

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\stock-opname-app.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\assets\*"; DestDir: "{app}\assets"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Stock Opname App"; Filename: "{app}\stock-opname-app.exe"
Name: "{group}\{cm:UninstallProgram,Stock Opname App}"; Filename: "{uninstallexe}"
Name: "{commondesktop}\Stock Opname App"; Filename: "{app}\stock-opname-app.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
