package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	al "github.com/libdns/alidns"
	"github.com/libdns/libdns"
)

func main() {
	accessKeyID := strings.TrimSpace(os.Getenv("ACCESS_KEY_ID"))
	accessKeySec := strings.TrimSpace(os.Getenv("ACCESS_KEY_SECRET"))
	
	if (accessKeyID == "") || (accessKeySec == "") {
		fmt.Printf("ERROR: %s\n", "ACCESS_KEY_ID or ACCESS_KEY_SECRET missing")
		return
	}

	zone := ""
	if len(os.Args) > 1 {
		zone = strings.TrimSpace(os.Args[1])
	}
	if zone == "" {
		fmt.Printf("ERROR: %s\n", "First arg zone missing")
		return
	}

	fmt.Printf("Get ACCESS_KEY_ID: %s,ACCESS_KEY_SECRET: %s,ZONE: %s\n", accessKeyID, accessKeySec, zone)
	provider := al.Provider{}
	provider.AccessKeyID = accessKeyID
	provider.AccessKeySecret = accessKeySec
	records, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}
	testID := ""
	testName := "_libdns_test"
	for _, record := range records {
		tmp := record.RR()
		fmt.Printf("%s (.%s): %s, %s\n", tmp.Name, zone, tmp.Data, tmp.Type)
		if testName == tmp.Name {
			r, _ := record.(al.DomainRecord)
			testID = r.ID
		}
	}
	if testID == "" {
		fmt.Println("Creating new entry for ", testName)
		records, err = provider.AppendRecords(context.TODO(), zone, []libdns.Record{al.DomainRecord{
			Type:  "TXT",
			Name:  testName,
			Value: fmt.Sprintf("This+is a test entry created by libdns %s", time.Now()),
			TTL:   600,
		}})
		if len(records) == 1 {
			tmp, _ := records[0].(al.DomainRecord)
			testID = tmp.ID
		}
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
	}

	fmt.Println("Press any Key to update the test entry")
	fmt.Scanln()
	if testID != "" {
		fmt.Println("Replacing entry for ", testName)
		records, err = provider.SetRecords(context.TODO(), zone, []libdns.Record{al.DomainRecord{
			Type:  "TXT",
			Name:  testName,
			Value: fmt.Sprintf("Replacement test entry upgraded by libdns %s", time.Now()),
			TTL:   605,
			ID:    testID,
		}})
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
	}
	fmt.Println("Press any Key to delete the test entry")
	fmt.Scanln()
	fmt.Println("Deleting the entry for ", testName)
	_, err = provider.DeleteRecords(context.TODO(), zone, records)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

}
