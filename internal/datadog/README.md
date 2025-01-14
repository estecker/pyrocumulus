
Response from `UsageMeteringApi.GetUsageBillableSummary`:
```json
{
  "org": {
    "datacenter": "us1.prod.dog"
  },
  "usage": [
    {
      "account_name": "Eddie",
      "account_public_id": "857dfafa-3998-45c4-b551-0deba5dc2fcc",
      "billing_plan": "Pro",
      "datacenter": "us1.prod.dog",
      "end_date": "2024-12-29T00:00:00Z",
      "num_orgs": 1,
      "org_name": "Eddie",
      "public_id": "857dfafa-3998-45c4-b551-0deba5dc2fcc",
      "ratio_in_month": 1,
      "region": "us",
      "start_date": "2024-12-01T00:00:00Z",
      "usage": {
        "apm_host_enterprise_sum": {
          "account_billable_usage": 273361,
          "billing_dimension": "apm_host_enterprise",
          "elapsed_usage_hours": 696,
          "first_billable_usage_hour": "2024-12-01T00:00:00+00:00",
          "last_billable_usage_hour": "2024-12-29T23:00:00+00:00",
          "org_billable_usage": 273361,
          "percentage_in_account": 100,
          "usage_unit": "host_hours"
        },
        "apm_trace_search_sum": {
          "account_billable_usage": 308770543,
          "billing_dimension": "apm_trace_search",
          "elapsed_usage_hours": 696,
          "first_billable_usage_hour": "2024-12-01T00:00:00Z",
          "last_billable_usage_hour": "2024-12-29T23:00:00Z",
          "org_billable_usage": 308770543,
          "percentage_in_account": 100,
          "usage_unit": "traces"
        },
        "timeseries_average": {
          "account_billable_usage": 322226,
          "billing_dimension": "timeseries",
          "elapsed_usage_hours": 696,
          "first_billable_usage_hour": "2024-12-01T00:00:00Z",
          "last_billable_usage_hour": "2024-12-29T23:00:00Z",
          "org_billable_usage": 322226,
          "percentage_in_account": 100,
          "usage_unit": "timeseries"
        }
      },
      "uuid": "857dfafa-3998-45c4-b551-0deba5dc2fcc"
    }
  ]
}
```