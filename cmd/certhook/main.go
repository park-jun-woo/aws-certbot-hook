// https://parkjunwoo.com/microstral/cmd/certhook/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"golang.org/x/net/publicsuffix"
)

func awsString(s string) *string { return &s }
func awsInt64(i int64) *int64    { return &i }

// findHostedZoneID는 certbot이 넘겨준 도메인에서
// 가장 잘 맞는 Route53 HostedZone ID를 찾아 리턴한다.
// 도메인 구조: foo.bar.domain.co.kr → HostedZone: domain.co.kr
func findHostedZoneID(r53 *route53.Client, domain string) (string, error) {
	// 1. publicsuffix를 통해 등록 가능한 루트 도메인(eTLD+1) 추출
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return "", fmt.Errorf("cannot parse public suffix: %w", err)
	}

	// 2. HostedZone 목록 전체 조회 (최대 100개 이상일 경우 paging 추가)
	var matchedZoneID string
	var maxLen int

	// Route53는 zone.Name이 항상 끝에 "." 붙음(정규화)
	listInput := &route53.ListHostedZonesInput{}
	for {
		out, err := r53.ListHostedZones(context.TODO(), listInput)
		if err != nil {
			return "", fmt.Errorf("failed to list hosted zones: %w", err)
		}
		for _, zone := range out.HostedZones {
			zoneName := strings.TrimSuffix(*zone.Name, ".")
			// eTLD+1 완전 일치 HostedZone이 있으면 바로 리턴 (가장 확실)
			if zoneName == rootDomain {
				return strings.TrimPrefix(*zone.Id, "/hostedzone/"), nil
			}
			// 아니면 suffix 매칭(서브도메인 지원, 길이가 긴 게 우선)
			if strings.HasSuffix(domain, zoneName) && len(zoneName) > maxLen {
				matchedZoneID = strings.TrimPrefix(*zone.Id, "/hostedzone/")
				maxLen = len(zoneName)
			}
		}
		if !out.IsTruncated {
			break
		}
		listInput.Marker = out.NextMarker
	}
	if matchedZoneID != "" {
		return matchedZoneID, nil
	}
	return "", fmt.Errorf("no matching hosted zone found for domain: %s", domain)
}

func main() {
	var hookType string
	var sleepTime int
	flag.StringVar(&hookType, "hook", "", "Hook type to run (auth|cleanup)")
	flag.IntVar(&sleepTime, "sleep", 20, "Sleep time after DNS update (seconds)")
	flag.Parse()

	domain := os.Getenv("CERTBOT_DOMAIN")
	validation := os.Getenv("CERTBOT_VALIDATION")
	if domain == "" || validation == "" {
		log.Fatalf("CERTBOT_DOMAIN or CERTBOT_VALIDATION is missing")
	}
	fqdn := "_acme-challenge." + domain + "."

	// AWS config 로드
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	r53 := route53.NewFromConfig(cfg)

	// HostedZoneId 자동 조회!
	hostedZoneID, err := findHostedZoneID(r53, domain)
	if err != nil || hostedZoneID == "" {
		log.Fatalf("failed to find matching Route53 HostedZoneId for %s: %v", domain, err)
	}
	log.Printf("Detected hostedZoneId=%s for domain=%s", hostedZoneID, domain)

	switch strings.ToLower(hookType) {
	case "auth":
		_, err := r53.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: &hostedZoneID,
			ChangeBatch: &types.ChangeBatch{
				Comment: awsString("certbot dns-01 auth hook"),
				Changes: []types.Change{
					{
						Action: types.ChangeActionUpsert,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: awsString(fqdn),
							Type: types.RRTypeTxt,
							TTL:  awsInt64(60),
							ResourceRecords: []types.ResourceRecord{
								{Value: awsString(`"` + validation + `"`)},
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Fatalf("failed to upsert TXT record: %v", err)
		}
		log.Printf("Successfully added TXT record %s", fqdn)
		time.Sleep(time.Duration(sleepTime) * time.Second)

	case "cleanup":
		_, err := r53.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: &hostedZoneID,
			ChangeBatch: &types.ChangeBatch{
				Comment: awsString("certbot dns-01 cleanup hook"),
				Changes: []types.Change{
					{
						Action: types.ChangeActionDelete,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: awsString(fqdn),
							Type: types.RRTypeTxt,
							TTL:  awsInt64(60),
							ResourceRecords: []types.ResourceRecord{
								{Value: awsString(`"` + validation + `"`)},
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Fatalf("failed to delete TXT record: %v", err)
		}
		log.Printf("Successfully deleted TXT record %s", fqdn)

	default:
		log.Fatalf("Unknown or missing --hook value: %s (must be one of: auth, cleanup)", hookType)
	}
}
