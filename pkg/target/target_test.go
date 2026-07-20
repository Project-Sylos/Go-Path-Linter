package target

import (
	"testing"

	"codeberg.org/Sylos/go-path-linter/pkg/check"
	"codeberg.org/Sylos/go-path-linter/pkg/issue"
)

func TestTargetsNonEmpty(t *testing.T) {
	cases := []struct {
		name string
		sep  string
		n    int
		got  string
	}{
		{"Windows", `\`, len(Windows().Checkers), Windows().Separator},
		{"MacOS", "/", len(MacOS().Checkers), MacOS().Separator},
		{"Linux", "/", len(Linux().Checkers), Linux().Separator},
		{"Dropbox", "/", len(Dropbox().Checkers), Dropbox().Separator},
		{"Box", "/", len(Box().Checkers), Box().Separator},
		{"Egnyte", "/", len(Egnyte().Checkers), Egnyte().Separator},
		{"OneDrive", "/", len(OneDrive().Checkers), OneDrive().Separator},
		{"SharePoint", "/", len(SharePoint().Checkers), SharePoint().Separator},
		{"ShareFile", "/", len(ShareFile().Checkers), ShareFile().Separator},
	}
	for _, tc := range cases {
		if tc.got != tc.sep {
			t.Fatalf("%s: sep want %q got %q", tc.name, tc.sep, tc.got)
		}
		if tc.n == 0 {
			t.Fatalf("%s: no checkers", tc.name)
		}
	}
}

func TestWindows_DriveRootAllowed(t *testing.T) {
	rs := Windows()
	var buf issue.Buffer
	rs.CheckAll([]string{"C:", "Users"}, check.CheckContext{Separator: `\`}, &buf)
	for _, iss := range buf.Issues {
		if iss.PartIndex == 0 && iss.Category == issue.CategoryInvalidChar {
			t.Fatalf("drive root must not be invalid-char: %#v", iss)
		}
	}
}

func TestWindows_ReservedDevice(t *testing.T) {
	rs := Windows()
	var buf issue.Buffer
	rs.CheckAll([]string{"C:", "CON"}, check.CheckContext{Separator: `\`}, &buf)
	found := false
	for _, iss := range buf.Issues {
		if iss.Category == issue.CategoryReservedName {
			found = true
		}
	}
	if !found {
		t.Fatal("expected reserved CON")
	}
}

func TestMacOS_ForceRelative(t *testing.T) {
	rs := MacOS()
	if !rs.ForceRelative || rs.Separator != "/" {
		t.Fatalf("macos policy: %#v", rs)
	}
}

func TestSharePoint_ComponentLimit(t *testing.T) {
	rs := SharePoint()
	b := make([]byte, 129)
	for i := range b {
		b[i] = 'a'
	}
	var buf issue.Buffer
	rs.CheckAll([]string{string(b)}, check.CheckContext{Separator: "/"}, &buf)
	found := false
	for _, iss := range buf.Issues {
		if iss.Category == issue.CategoryLength {
			found = true
		}
	}
	if !found {
		t.Fatal("sharepoint component max is 128")
	}
}

func TestOneDrive_ReservedForms(t *testing.T) {
	rs := OneDrive()
	cleaned, actions := rs.CleanAll([]string{"_vti_"}, check.CheckContext{Separator: "/"})
	if cleaned[0] == "_vti_" {
		t.Fatalf("expected rename, actions=%v", actions)
	}
}
