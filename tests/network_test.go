package web_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/defos-ai/commander-web-agent/actions"
	"golang.org/x/net/dns/dnsmessage"
)

func TestWebsiteRejectsReservedAddresses(t *testing.T) {
	for _, address := range []string{
		"169.254.169.254", "169.254.170.2", "100.100.100.200", "168.63.129.16",
		"fd00:ec2::254", "::ffff:169.254.169.254", "fe80::1", "fe80::1%lo",
		"0.0.0.0", "::", "224.0.0.1", "ff02::1", "240.0.0.1", "192.0.2.1",
		"198.18.0.1", "2001:db8::1", "64:ff9b::a9fe:a9fe", "2002:a9fe:a9fe::1",
	} {
		t.Run(address, func(t *testing.T) {
			target := (&url.URL{Scheme: "http", Host: net.JoinHostPort(address, "80")}).String()
			raw, _ := json.Marshal(actions.PageInput{URL: target})
			executor := actions.NewCheckLinks()
			if _, err := executor.Discover(t.Context(), nil, nil); err != nil {
				t.Fatal(err)
			}
			_, err := executor.Run(t.Context(), raw)
			if err == nil || !strings.Contains(err.Error(), "blocked") {
				t.Fatalf("reserved address was not blocked: %v", err)
			}
		})
	}
}

func TestWebsitePinsDNSAcrossRequestsAndPortChanges(t *testing.T) {
	var lookups atomic.Int32
	withWebsiteDNS(t, func() []netip.Addr {
		if lookups.Add(1) == 1 {
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}
		}
		return []netip.Addr{netip.MustParseAddr("169.254.169.254")}
	})
	about := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "<title>About</title>")
	}))
	defer about.Close()
	home := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `<title>Home</title><a href="%s">About</a>`, websiteURL(about.URL))
	}))
	defer home.Close()
	result := runAction(t, actions.NewDiscoverPagesNoBrowser(), fmt.Sprintf(`{"url":%q,"max_pages":2}`, websiteURL(home.URL)))
	var pages actions.DiscoverPagesOutput
	if err := json.Unmarshal(result.Output, &pages); err != nil {
		t.Fatal(err)
	}
	if len(pages.Pages) != 2 || lookups.Load() != 1 {
		t.Fatalf("DNS was not pinned across page requests and port changes: lookups=%d output=%s", lookups.Load(), result.Output)
	}
}

func TestWebsiteRejectsMixedDNSAnswers(t *testing.T) {
	withWebsiteDNS(t, func() []netip.Addr {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("169.254.169.254")}
	})
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("connected despite an unsafe DNS answer")
	}))
	defer site.Close()
	executor := actions.NewCheckLinks()
	if _, err := executor.Discover(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(actions.PageInput{URL: websiteURL(site.URL)})
	if _, err := executor.Run(t.Context(), raw); err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("mixed DNS answer accepted: %v", err)
	}
}

func TestWebsiteIgnoresEnvironmentProxy(t *testing.T) {
	withWebsiteDNS(t, func() []netip.Addr { return []netip.Addr{netip.MustParseAddr("127.0.0.1")} })
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("used environment proxy")
	}))
	defer proxy.Close()
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "<title>Site</title>")
	}))
	defer site.Close()
	runAction(t, actions.NewCheckLinks(), fmt.Sprintf(`{"url":%q}`, websiteURL(site.URL)))
}

func websiteURL(raw string) string {
	u, _ := url.Parse(raw)
	u.Host = net.JoinHostPort("website.test", u.Port())
	return u.String()
}

// Exercise the public action with changing DNS answers; never resolve a public
// hostname or change production policy just to inject a test resolver.
func withWebsiteDNS(t *testing.T, answers func() []netip.Addr) {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 4096)
		for {
			n, address, err := conn.ReadFrom(buffer)
			if err != nil {
				return
			}
			var query dnsmessage.Message
			if err := query.Unpack(buffer[:n]); err != nil {
				t.Errorf("decode DNS request: %v", err)
				continue
			}
			response := dnsmessage.Message{Header: dnsmessage.Header{ID: query.ID, Response: true, RecursionDesired: query.RecursionDesired, RecursionAvailable: true}, Questions: query.Questions}
			for _, question := range query.Questions {
				if question.Name.String() != "website.test." {
					t.Errorf("unexpected DNS name: %s", question.Name)
					continue
				}
				if question.Type != dnsmessage.TypeA {
					continue
				}
				for _, ip := range answers() {
					response.Answers = append(response.Answers, dnsmessage.Resource{
						Header: dnsmessage.ResourceHeader{Name: question.Name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET},
						Body:   &dnsmessage.AResource{A: ip.As4()},
					})
				}
			}
			data, err := response.Pack()
			if err != nil {
				t.Errorf("encode DNS response: %v", err)
				continue
			}
			if _, err := conn.WriteTo(data, address); err != nil {
				return
			}
		}
	}()
	original := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "udp4", conn.LocalAddr().String())
	}}
	t.Cleanup(func() { net.DefaultResolver = original; conn.Close(); <-done })
}
