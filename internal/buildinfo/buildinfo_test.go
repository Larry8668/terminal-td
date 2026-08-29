package buildinfo

import "testing"

func TestIsHomebrewChannelOverride(t *testing.T) {
	orig := Channel
	defer func() { Channel = orig }()

	Channel = "homebrew"
	if !IsHomebrew("/usr/local/bin/terminal-td") {
		t.Fatal("expected Channel=homebrew to report IsHomebrew=true regardless of path")
	}
}

func TestIsHomebrewPathHeuristic(t *testing.T) {
	orig := Channel
	defer func() { Channel = orig }()
	Channel = "source"

	cases := map[string]bool{
		"/opt/homebrew/Cellar/terminal-td/1.2.3/bin/terminal-td": true,
		"/usr/local/Caskroom/terminal-td/1.2.3/terminal-td":      true,
		`C:\Program Files\Caskroom\terminal-td\terminal-td.exe`:  true,
		"/Users/me/project/terminal-td/builds/terminal-td":       false,
		"/home/me/go/bin/terminal-td":                            false,
	}
	for path, want := range cases {
		if got := IsHomebrew(path); got != want {
			t.Errorf("IsHomebrew(%q) = %v, want %v", path, got, want)
		}
	}
}
