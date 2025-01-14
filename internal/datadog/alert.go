package datadog

//
//import (
//	"fmt"
//	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
//	"os"
//	"time"
//)
//
//func GetHourlyUsage(ctx context.Context, api *datadogV2.UsageMeteringApi) {
//resp, r, err := api.GetHourlyUsage(ctx, time.Now().AddDate(0, 0, -2), "infra_hosts", *datadogV2.NewGetHourlyUsageOptionalParameters().WithFilterTimestampEnd(time.Now().AddDate(0, 0, 1)))
//
//if err != nil {
//fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetHourlyUsage`: %v\n", err)
//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
//}
//
//responseContent, _ := json.MarshalIndent(resp, "", "  ")
//fmt.Fprintf(os.Stdout, "Response from `UsageMeteringApi.GetHourlyUsage`:\n%s\n", responseContent)
//}
