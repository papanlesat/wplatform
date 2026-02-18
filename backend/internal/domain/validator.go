package domain

import (
	"context"
	"fmt"
	"net"
)

type DomainValidator interface {
	VerifyDomainOwnership(ctx context.Context, domain string, serverIP string) error
	ValidateDNSRecord(ctx context.Context, domain string, expectedIP string) error
}

type DNSValidator struct {
	dnsResolver string
}

func NewDNSValidator() *DNSValidator {
	return &DNSValidator{
		dnsResolver: "8.8.8.8",
	}
}

func NewDNSValidatorWithResolver(dnsResolver string) *DNSValidator {
	if dnsResolver == "" {
		dnsResolver = "8.8.8.8"
	}
	return &DNSValidator{
		dnsResolver: dnsResolver,
	}
}

func (dv *DNSValidator) VerifyDomainOwnership(ctx context.Context, domain string, serverIP string) error {
	records, err := dv.resolveDNSRecords(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to resolve DNS records for %s: %w", domain, err)
	}

	for _, record := range records {
		if record == serverIP {
			return nil
		}
	}

	return fmt.Errorf("domain %s does not point to server IP %s", domain, serverIP)
}

func (dv *DNSValidator) ValidateDNSRecord(ctx context.Context, domain string, expectedIP string) error {
	records, err := dv.resolveDNSRecords(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to resolve DNS records: %w", err)
	}

	for _, record := range records {
		if record == expectedIP {
			return nil
		}
	}

	return fmt.Errorf("DNS A record for %s does not match expected IP %s", domain, expectedIP)
}

func (dv *DNSValidator) resolveDNSRecords(ctx context.Context, domain string) ([]string, error) {
	var records []string

	cnameRecords, err := net.LookupCNAME(domain + ".")
	if err == nil {
		if len(cnameRecords) > 0 {
			return nil, fmt.Errorf("domain %s is a CNAME, not an A record", domain, cnameRecords[0])
		}
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed for %s: %w", domain, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no DNS records found for domain %s", domain)
	}

	for _, ip := range ips {
		records = append(records, ip.String())
	}

	return records, nil
}

func (dv *DNSValidator) GetServerIPAddress() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", fmt.Errorf("failed to create DNS connection: %w", err)
	}
	defer conn.Close()

	addresses, err := net.LookupIP("resolver.opendns.com")
	if err != nil {
		return "", fmt.Errorf("failed to resolve server IP: %w", err)
	}

	if len(addresses) == 0 {
		return "", fmt.Errorf("no external IP addresses found")
	}

	return addresses[0].String(), nil
}
