# LOCAL LIVE SOURCE CONFIGURATION

Updated: 2026-09-19

The repository is public. Production endpoint URLs, request bodies, credentials, tokens and company data must never be committed to GitHub.

## Owner workflow

A separate runtime-profile file is **no longer required** for normal operation.

The application exposes two business data sources in **Thiết lập**:

1. **Sản lượng**
   - source used for Payroll/Productivity;
   - parsed into User/PDA;
   - processed into Phân ca;
   - processed into Pick and Pack.

2. **Đang lấy hàng**
   - source used for Active Picking;
   - parsed into the live progress/status table.

For each source, the Owner performs the setup once:

1. choose the source in the application;
2. paste the matching Dashboard cURL (bash);
3. click **LƯU NGUỒN**;
4. click **KIỂM TRA NGUỒN**.

The application then:
- extracts URL, HTTP method, body and non-sensitive request headers;
- extracts Dashboard session values separately;
- removes the raw cURL from the edit box;
- protects the session and source configuration with Windows DPAPI for the current Windows user;
- keeps the source configuration outside the public repository and release artifact.

After both sources are configured, normal operation only requires **ĐỒNG BỘ**.

## Session refresh

Dashboard credentials may expire before the saved request shape changes.

Pasting a fresh cURL for either configured source:
- refreshes non-empty session values;
- refreshes that source request definition;
- preserves the other configured source.

## Daily request dates

When a saved cURL contains the current, previous or next calendar date, the application converts those values into local templates:
- `{{YESTERDAY_ISO}}`, `{{TODAY_ISO}}`, `{{TOMORROW_ISO}}`;
- `{{YESTERDAY_DMY}}`, `{{TODAY_DMY}}`, `{{TOMORROW_DMY}}`.

The values are expanded at synchronization time so a request captured today does not stay pinned to an old date tomorrow.

## Synchronization pipeline

### Sản lượng
Dashboard request
→ XLSX response
→ header/field detection
→ completed-work rows
→ User/PDA
→ automatic/manual shift resolution
→ within-shift/overtime split
→ even/odd aggregation
→ Pick / Pack / Phân ca tables.

### Đang lấy hàng
Dashboard request
→ JSON response
→ Active Picking field mapping
→ status/warning/progress/time/device fields
→ Đang lấy hàng table.

## Partial failure behavior

The two sources are independent.

If one source fails:
- the successful source is still processed;
- the last valid data from the failed source is retained;
- status/log identifies the missing/failed business source;
- the app does not clear valid data only because another source failed.

## Security

The local encrypted source/session files may contain operational request information and must not be uploaded to:
- public GitHub issues;
- commits;
- releases;
- Actions artifacts;
- public diagnostics.

The exported application Log remains sanitized and does not include raw endpoint URLs or credentials.

## Legacy compatibility

The old `SupraProductivity.profile.json` import path remains only as backward compatibility for existing local installations. New Owner setup must use the in-app source configuration workflow.
