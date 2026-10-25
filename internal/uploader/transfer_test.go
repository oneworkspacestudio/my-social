package uploader

import (
	"net/url"
	"strings"
	"testing"
)

func TestLaunchDestinationAndCredentialValidation(t *testing.T) {
	ticket := strings.Repeat("a", 48) + "." + strings.Repeat("b", 48)
	for _, tc := range []struct {
		base string
		ok   bool
	}{{"https://social.example/social", true}, {"http://localhost:8123", true}, {"http://127.0.0.1:8123", true}, {"http://social.example/social", false}, {"https://evil@example.test/social", false}, {"https://social.example/drive", false}, {"https://social.example/social?target=x", false}, {"https://social.example/social#x", false}, {"https://social.example/social/../drive", false}, {"https://social.example/%73ocial", false}} {
		launch := "mysocial-upload://start/?" + url.Values{"server": {tc.base}, "ticket": {ticket}, "account": {strings.Repeat("c", 64)}}.Encode()
		s, e := ParseLaunch(launch)
		if (e == nil) != tc.ok {
			t.Errorf("destination %s: %v", tc.base, e)
		}
		if e == nil && (!isSecret(s.Client) || s.Status != "queued") {
			t.Fatal("missing independent client credential")
		}
	}
	for _, launch := range []string{"https://evil.example", `mysocial-upload://start/?server=https://social.example/social&ticket=bad`, "mysocial-upload://start/?server=https://social.example/social&ticket=" + ticket + "&extra=x"} {
		if _, e := ParseLaunch(launch); e == nil {
			t.Fatal("malformed launch accepted")
		}
	}
}

func TestInteractiveHandoffCannotTrustAnImpersonatedLoopbackServer(t *testing.T) {
	for _, base := range []string{"http://localhost:8123", "http://127.0.0.1:8123", "http://localhost:8123/social"} {
		if InteractiveBase(State{Base: base}) == nil {
			t.Fatal("interactive native handoff accepted insecure local identity", base)
		}
	}
	if e := InteractiveBase(State{Base: "https://social.example/social"}); e != nil {
		t.Fatal(e)
	}
}
