package version

import "testing"

func TestCurrentIdentityAndCLIFormat(t *testing.T) {
	i := Current()
	if i.Name != ApplicationName || i.Version != ApplicationVersion {
		t.Fatalf("identity=%+v", i)
	}
	if got := CLI("musicprobe"); got != "MusicUnlocker musicprobe 0.7.0 (build dev, built unknown)" {
		t.Fatalf("CLI=%q", got)
	}
}
