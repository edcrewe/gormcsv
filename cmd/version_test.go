package cmd

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	revision := []debug.BuildSetting{{Key: "vcs.revision", Value: "a6866f6602fbedb5007c63d9579c1218fbbd573f"}}
	tests := []struct {
		name   string
		linked string
		info   *debug.BuildInfo
		ok     bool
		want   string
	}{
		{name: "linked tag", linked: "v0.3.1", ok: true, want: "v0.3.1"},
		{name: "linked describe", linked: "v0.3.1-2-ga6866f6", ok: true, want: "v0.3.1-2-ga6866f6"},
		{
			name: "linked wins over build info", linked: "v0.3.1",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, ok: true, want: "v0.3.1",
		},
		{name: "go install at tag", info: &debug.BuildInfo{Main: debug.Module{Version: "v0.3.1"}}, ok: true, want: "v0.3.1"},
		{
			name: "go build pseudo-version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v0.3.2-0.20261113094732-a6866f6602fb"}},
			ok:   true, want: "v0.3.2-0.20261113094732-a6866f6602fb",
		},
		{
			name: "devel with revision",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: revision},
			ok:   true, want: "devel-a6866f6",
		},
		{name: "devel without vcs", info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, ok: true, want: "devel"},
		{name: "no build info", ok: false, want: "devel"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveVersion(test.linked, test.info, test.ok); got != test.want {
				t.Errorf("resolveVersion() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"version"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "gormcsv "+currentVersion()+"\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func TestVersionCommandRejectsArgs(t *testing.T) {
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"version", "extra"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err == nil {
		t.Error("expected an error for unexpected arguments")
	}
}
