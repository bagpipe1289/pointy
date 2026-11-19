package ui

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/pointylab/pointy/desktop/internal/config"
)

// fakeAgent stands in for the DSH relay: it records what a closed card asks
// of it, and says whether the plugin is there to hear.
type fakeAgent struct {
	archived  []string
	connected bool
}

func (f *fakeAgent) Request(jsontext.Value) bool { return f.connected }
func (f *fakeAgent) Open(string)                 {}
func (f *fakeAgent) Archive(sessionID string) bool {
	f.archived = append(f.archived, sessionID)
	return f.connected
}

func TestClosingACardArchivesTheSessionItAskedIn(t *testing.T) {
	relay := &fakeAgent{connected: true}
	settings := config.LoadSettings(t.TempDir())
	u := &UI{deps: Deps{Agent: relay, Settings: settings}}

	u.archiveSession("s1")
	if len(relay.archived) != 1 || relay.archived[0] != "s1" {
		t.Fatalf("archived %v", relay.archived)
	}
	u.archiveSession("")
	if len(relay.archived) != 1 {
		t.Fatal("archived a card that asked in no session")
	}

	if err := settings.Update(func(s *config.Settings) { s.AutoArchive = false }); err != nil {
		t.Fatal(err)
	}
	u.archiveSession("s2")
	if len(relay.archived) != 1 {
		t.Fatal("archived with auto-archive turned off in the tray")
	}
}

func TestArchiveSaysSoWhenThePluginIsGone(t *testing.T) {
	var logged []string
	u := &UI{deps: Deps{Agent: &fakeAgent{}, Settings: config.LoadSettings(t.TempDir()),
		Log: func(kind string, _ any) { logged = append(logged, kind) }}}
	u.archiveSession("s1")
	if len(logged) != 1 || logged[0] != "archive" {
		t.Fatalf("logged %v", logged)
	}
}

func TestAutoArchiveIsOnWithoutSettings(t *testing.T) {
	if !autoArchives(nil) {
		t.Fatal("auto-archive off without settings")
	}
}

func TestArchiveWithoutARelayKeepsQuiet(t *testing.T) {
	u := &UI{deps: Deps{Settings: config.LoadSettings(t.TempDir())}}
	u.archiveSession("s1")
}

func TestArchiveHintShowsOnce(t *testing.T) {
	settings := config.LoadSettings(t.TempDir())
	u := &UI{deps: Deps{Settings: settings}}
	if text := u.archiveHint("s1"); text == "" {
		t.Fatal("no hint for the first archived session")
	}
	if !settings.Get().ArchiveHinted {
		t.Fatal("the hint shown was not recorded")
	}
	if text := u.archiveHint("s2"); text != "" {
		t.Fatalf("hint shown twice: %q", text)
	}
	if text := u.archiveHint(""); text != "" {
		t.Fatal("hinted without a session")
	}
	if text := (&UI{}).archiveHint("s3"); text != "" {
		t.Fatal("hinted without settings")
	}
}
