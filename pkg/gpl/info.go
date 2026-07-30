package gpl

// Info is display metadata for a built-in target (UI / docs links).
type Info struct {
	ID          string
	DisplayName string
	DocsURL     string
}

// InfoFor returns naming-docs metadata for a built-in target.
func InfoFor(t Target) Info {
	switch t {
	case Windows:
		return Info{
			ID:          "windows",
			DisplayName: "Windows",
			DocsURL:     "https://learn.microsoft.com/windows/win32/fileio/naming-a-file",
		}
	case MacOS:
		return Info{
			ID:          "macos",
			DisplayName: "macOS",
			DocsURL:     "https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemDetails/FileSystemDetails.html",
		}
	case Linux:
		return Info{
			ID:          "linux",
			DisplayName: "Linux",
		}
	case Dropbox:
		return Info{
			ID:          "dropbox",
			DisplayName: "Dropbox",
			DocsURL:     "https://help.dropbox.com/organize/file-names",
		}
	case Box:
		return Info{
			ID:          "box",
			DisplayName: "Box",
			DocsURL:     "https://support.box.com/hc/en-us/articles/360043696410-File-and-folder-name-restrictions",
		}
	case Egnyte:
		return Info{
			ID:          "egnyte",
			DisplayName: "Egnyte",
			DocsURL:     "https://helpdesk.egnyte.com/hc/en-us/articles/201637704-File-and-Folder-Name-Restrictions",
		}
	case OneDrive:
		return Info{
			ID:          "onedrive",
			DisplayName: "OneDrive",
			DocsURL:     "https://support.microsoft.com/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb94e041e1e1",
		}
	case SharePoint:
		return Info{
			ID:          "sharepoint",
			DisplayName: "SharePoint",
			DocsURL:     "https://support.microsoft.com/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb94e041e1e1",
		}
	case ShareFile:
		return Info{
			ID:          "sharefile",
			DisplayName: "ShareFile",
			DocsURL:     "https://docs.sharefile.com/en-us/sharefile/admin/configure/advanced-preferences/file-and-folder-name-restrictions.html",
		}
	default:
		return Info{ID: string(t), DisplayName: string(t)}
	}
}
