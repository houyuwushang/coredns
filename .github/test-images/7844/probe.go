package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/miekg/dns"
)

func main() {
	protocol := flag.String("net", "udp", "DNS transport")
	flag.Parse()
	if err := probe(*protocol); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func probe(protocol string) error {
	client := &dns.Client{Net: protocol, Timeout: 3 * time.Second}
	for n := range 3 {
		question := new(dns.Msg)
		question.SetQuestion("example.test.", dns.TypeA)
		answer, elapsed, err := client.Exchange(question, "127.0.0.1:53")
		if err != nil {
			return fmt.Errorf("%s query %d: %w", protocol, n+1, err)
		}
		if answer.Rcode != dns.RcodeSuccess || len(answer.Answer) != 1 {
			return fmt.Errorf("%s query %d: unexpected response: %s", protocol, n+1, answer)
		}
		record, ok := answer.Answer[0].(*dns.A)
		if !ok || record.A.String() != "203.0.113.53" {
			return fmt.Errorf("%s query %d: unexpected answer: %s", protocol, n+1, answer)
		}
		fmt.Printf("%s query %d: NOERROR, %s\n", protocol, n+1, elapsed)
	}
	return nil
}
