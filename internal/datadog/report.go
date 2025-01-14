package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"pyrocumulus/internal/common"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/fatih/structs"
	"os"
)

func Report(tagBreakdownKeys string, costAttributionKey string) {
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: os.Getenv("DD_API_KEY"),
			},
			"appKeyAuth": {
				Key: os.Getenv("DD_APP_KEY"),
			},
		},
	)

	configuration := datadog.NewConfiguration()
	configuration.RetryConfiguration.EnableRetry = true
	configuration.SetUnstableOperationEnabled("v2.GetMonthlyCostAttribution", true)
	configuration.SetUnstableOperationEnabled("v2.GetActiveBillingDimensions", true)
	configuration.SetUnstableOperationEnabled("v2.GetMonthlyCostAttribution", true)
	apiClient := datadog.NewAPIClient(configuration)

	apiv2 := datadogV2.NewUsageMeteringApi(apiClient)
	apiv1 := datadogV1.NewUsageMeteringApi(apiClient)

	rsw := viper.GetString("reports-slack-webhook")
	if rsw != "" {
		_ = slack.Send(rsw, "", slack.Payload{Text: GetEstimatedCostByOrg(ctx, apiv2)})
		_ = slack.Send(rsw, "", slack.Payload{Text: GetProjectedCost(ctx, apiv2)})
		_ = slack.Send(rsw, "", slack.Payload{Text: GetUnTagged(ctx, apiv1, tagBreakdownKeys)})
		_ = slack.Send(rsw, "", slack.Payload{Text: GetUsageBillableSummary(ctx, apiv1)})
		_ = slack.Send(rsw, "", slack.Payload{Text: "More details can be found at https://app.datadoghq.com/billing/usage"})
	} else {
		GetEstimatedCostByOrg(ctx, apiv2)
		GetProjectedCost(ctx, apiv2)
		GetUnTagged(ctx, apiv1, tagBreakdownKeys)
		GetUsageBillableSummary(ctx, apiv1)
	}
}

// Get estimated cost across multi-org and single root-org accounts. Estimated cost data is only available for
// the current month and previous month and is delayed by up to 72 hours from when it was incurred.
// To access historical costs prior to this, use the /historical_cost endpoint.
//
//	https://api.datadoghq.com/api/v2/usage/estimated_cost
func GetEstimatedCostByOrg(ctx context.Context, api *datadogV2.UsageMeteringApi) string {
	resp, r, err := api.GetEstimatedCostByOrg(ctx, *datadogV2.NewGetEstimatedCostByOrgOptionalParameters().WithStartMonth(time.Now()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetEstimatedCostByOrg`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return ""
	}
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	p := message.NewPrinter(language.English)
	top := 5
	for _, pc := range resp.Data {
		p.Fprintf(&sb, "%s Datadog monthly total cost for %s is: $%.2f\n", pc.Attributes.GetOrgName(), pc.Attributes.GetDate().Format("2006-01"), pc.Attributes.GetTotalCost())
		fmt.Fprintf(&sb, "Top %d total costs are:", top)
		sort.Slice(pc.Attributes.Charges, func(i, j int) bool {
			return pc.Attributes.Charges[i].GetCost() > pc.Attributes.Charges[j].GetCost()
		})
		fmt.Fprintf(tw, "%s", "```\n")
		for _, attribute := range pc.Attributes.Charges {
			if top > 0 && attribute.GetChargeType() == "total" {
				p.Fprintf(tw, "%s:\t\t$%.2f\n", attribute.GetProductName(), attribute.GetCost())
				top--
			}
		}
		fmt.Fprintf(tw, "%s", "```")
	}
	tw.Flush()
	fmt.Println(sb.String())
	return sb.String()
}

// Get projected cost across multi-org and single root-org accounts. Projected cost data is only available for the current month and becomes available around the 12th of the month.
//
//	https://api.datadoghq.com/api/v2/usage/projected_cost
func GetProjectedCost(ctx context.Context, api *datadogV2.UsageMeteringApi) string {
	resp, r, err := api.GetProjectedCost(ctx, *datadogV2.NewGetProjectedCostOptionalParameters().WithIncludeConnectedAccounts(true))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetProjectedCost`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return ""
	}
	var sb strings.Builder
	p := message.NewPrinter(language.English)
	for _, pc := range resp.Data {
		if pc.Attributes.GetProjectedTotalCost() == 0 { // Projected cost data is only available for the current month and becomes available around the 12th of the month.
			p.Fprintf(&sb, "Datadog: No total monthly projected costs for %s as of %s\n", pc.Attributes.GetOrgName(), pc.Attributes.GetDate().Format("2006-01-02"))
		} else {
			p.Fprintf(&sb, "Datadog: Monthly Datadog projected total costs for %s as of %s is: $%.2f\n", pc.Attributes.GetOrgName(), pc.Attributes.GetDate().Format("2006-01-02"), pc.Attributes.GetProjectedTotalCost())
		}
	}
	fmt.Println(sb.String())
	return sb.String()
}

// Action item: How much of the usage is untagged?
// https://docs.datadoghq.com/api/latest/usage-metering/#get-monthly-cost-attribution
func GetUnTagged(ctx context.Context, api *datadogV1.UsageMeteringApi, tagBreakdownKeys string) string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	resp, r, err := api.GetMonthlyUsageAttribution(ctx, time.Now(), datadogV1.MONTHLYUSAGEATTRIBUTIONSUPPORTEDMETRICS_ALL, *datadogV1.NewGetMonthlyUsageAttributionOptionalParameters().WithTagBreakdownKeys(tagBreakdownKeys))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetUnTagged`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return ""
	}
	for _, usageByTags := range resp.Usage {
		uTags := usageByTags.Tags //a map where the keys are strings and the values are slices of strings.
		allEmpty := true
		for _, values := range uTags {
			if len(values) != 0 {
				allEmpty = false
				break
			}
		}
		if allEmpty { // Only way I could figure out how to get all k/v's without calling all the methods
			fmt.Fprintf(&sb, "Datadog: The percentage of usage that is missing all the \"%s\" tags, greater than 0%%, for the month of %s in %s is:", tagBreakdownKeys, usageByTags.Month.Month(), usageByTags.GetOrgName())
			var usage map[string]float64
			valuesJSON, _ := json.Marshal(usageByTags.Values)
			json.Unmarshal(valuesJSON, &usage)
			fmt.Fprintf(tw, "%s", "```\n")
			for key, value := range usage {
				if value > 0 && strings.Contains(key, "_percentage") {
					fmt.Fprintf(tw, "%s:\t%.2f%%\n", key, value)
				}
			}
		}
	}
	fmt.Fprintf(tw, "%s", "```")
	tw.Flush()
	fmt.Println(sb.String())
	return sb.String()
}

// Compared to last month: How are you doing?
func GetUsageBillableSummary(ctx context.Context, api *datadogV1.UsageMeteringApi) string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	lm := getUsageBillableSummary(ctx, api, time.Now().AddDate(0, -1, 0))
	tm := getUsageBillableSummary(ctx, api, time.Now())

	for org, usageKey := range tm {
		fmt.Fprintf(&sb, "Datadog: Percentage difference of usage for %s between last month and this month. ", org)
		fmt.Fprintf(&sb, "Current month of %s is %d%% complete.", time.Now().Month().String(), common.PercentageOfMonth())
		fmt.Fprintf(tw, "%s", "```\n")
		for key, usage := range usageKey {
			var stat string
			percentDiff := float64(usage.AccountBillableUsage) / float64(lm[org][key].AccountBillableUsage) * 100
			if int(percentDiff) <= common.PercentageOfMonth() {
				stat = "🥳"
			} else {
				stat = "💸"
			}
			fmt.Fprintf(tw, "%s:\t\t%.f%%\t%s\t%s\n", key, percentDiff, usage.UsageUnit, stat)
		}
		fmt.Fprintf(tw, "%s", "```")
	}
	tw.Flush()
	fmt.Println(sb.String())
	return sb.String()
}

// https://api.datadoghq.com/api/v1/usage/billable-summary
// Painful process to get correct data from the API. Golang library likes to stuff data into the AdditionalProperties field.
func getUsageBillableSummary(ctx context.Context, api *datadogV1.UsageMeteringApi, month time.Time) map[string]map[string]usageBillableSummaryBody {
	var sb strings.Builder
	ubsb := make(map[string]map[string]usageBillableSummaryBody) // {Org: {UsageBillableSummaryKeys: {usageBillableSummaryBody}}, ...}]
	usageBillableSummary, r, err := api.GetUsageBillableSummary(ctx, *datadogV1.NewGetUsageBillableSummaryOptionalParameters().WithMonth(month).WithIncludeConnectedAccounts(true))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetUsageBillableSummary`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil
	}
	for _, org := range usageBillableSummary.GetUsage() { //datadogV1.UsageBillableSummaryHour:= UsageBillableSummaryResponse
		if ubsb[org.GetOrgName()] == nil {
			ubsb[org.GetOrgName()] = make(map[string]usageBillableSummaryBody)
		}
		usage := org.GetUsage() //datadogV1.UsageBillableSummaryKeys

		//fmt.Fprintf(&sb, "Org: %s, Start Date: %s, End Date: %s \n", org.GetOrgName(), org.GetStartDate().Format("2006-01-02"), org.GetEndDate().Format("2006-01-02"))
		fields := structs.Fields(usage) //So I can iterate through the fields
		for _, value := range fields {
			if !value.IsZero() {
				switch v := value.Value().(type) {
				case map[string]interface{}: // AdditionalProperties end up here.
					for _, y := range v {
						aps, ok := y.(map[string]interface{})
						if ok {
							//fmt.Fprintf(&sb, "%s: %.0f %s \n", aps["billing_dimension"], aps["account_billable_usage"], aps["usage_unit"])
							ubsb[org.GetOrgName()][aps["billing_dimension"].(string)] = usageBillableSummaryBody{
								AccountBillableUsage: int64(aps["account_billable_usage"].(float64)),
								ElapsedUsageHours:    int64(aps["elapsed_usage_hours"].(float64)),
								OrgBillableUsage:     int64(aps["org_billable_usage"].(float64)),
								PercentageInAccount:  aps["percentage_in_account"].(float64),
								UsageUnit:            aps["usage_unit"].(string),
								BillingDimension:     aps["billing_dimension"].(string),
							}
						} else {
							fmt.Fprintf(&sb, "Technical difficulties trying to casting interface to a type %T\n", v)
						}
					}
				case *datadogV1.UsageBillableSummaryBody:
					//	fmt.Fprintf(&sb, "%s: %d %s\n", v.AdditionalProperties["billing_dimension"], v.GetAccountBillableUsage(), v.GetUsageUnit())
					ubsb[org.GetOrgName()][v.AdditionalProperties["billing_dimension"].(string)] = usageBillableSummaryBody{
						AccountBillableUsage: v.GetAccountBillableUsage(),
						ElapsedUsageHours:    v.GetElapsedUsageHours(),
						OrgBillableUsage:     v.GetOrgBillableUsage(),
						PercentageInAccount:  v.GetPercentageInAccount(),
						UsageUnit:            v.GetUsageUnit(),
						BillingDimension:     v.AdditionalProperties["billing_dimension"].(string),
					}
				default:
					panic(fmt.Sprintf("Unknown type %T\n", v)) // Should never happen
				}
			}
		}
	}
	fmt.Println(sb.String())
	return ubsb
}

// My own struct to hold the data from `UsageMeteringApi.GetUsageBillableSummary`
// The datadog-api-client-go library likes to cheat and stuff data into the AdditionalProperties field.
type usageBillableSummaryBody struct {
	AccountBillableUsage int64
	ElapsedUsageHours    int64
	OrgBillableUsage     int64
	PercentageInAccount  float64
	UsageUnit            string
	BillingDimension     string
}
